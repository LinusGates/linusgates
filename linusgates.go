package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/LinusGates/osmgr"
	"github.com/LinusGates/reghive"
	"github.com/otiai10/copy"
)

func hookInstallSelectWim(me *MenuEngine) {
	menu := me.Menus["install-select_wim"]
	menu.Items = make([]*MenuItem, 0)

	for i := 0; i < len(SOURCES); i++ {
		editions := ""
		for j := 0; j < len(SOURCES[i].IMAGE); j++ {
			if editions != "" {
				editions += "\n - "
			}
			editions += SOURCES[i].IMAGE[j].Summary
		}
		menu.AddItem(filepath.Base(SOURCES[i].Path), editions, "setvar WIM "+SOURCES[i].Path, "menu install-select_edition")
	}
	me.Redraw()
}

func hookInstallSelectEdition(me *MenuEngine) {
	menu := me.Menus["install-select_edition"]
	menu.Items = make([]*MenuItem, 0)

	if wim, ok := me.Environment["WIM"]; ok && wim != "" {
		wiminfo, err := WIMInfo(wim)
		if err != nil {
			delete(me.Environment, "WIM")
			menu.Title = "Failed to verify WIM!"
			menu.Subtitle = err.Error()
			return
		}

		if len(wiminfo.IMAGE) == 0 {
			delete(me.Environment, "WIM")
			menu.Title = "WIM verified, but has no editions!"
			return
		}
		SOURCEIMG = wiminfo

		for i := 0; i < len(wiminfo.IMAGE); i++ {
			wimimage := wiminfo.IMAGE[i]
			menu.AddItem(wimimage.INDEX+". "+wimimage.NAME.Text, wimimage.Summary, "setvar INDEX "+wimimage.INDEX, "menu install-select_disk")
		}
	} else {
		me.ErrorText("No WIM was specified!", "")
	}
	me.Redraw()
}

func hookInstallSelectDisk(me *MenuEngine) {
	menu := me.Menus["install-select_disk"]
	menu.Items = make([]*MenuItem, 0)

	//We select an edition before reaching here, so set our environment now
	wimindex, err := strconv.Atoi(me.Environment["INDEX"])
	if err != nil {
		menu.Title = "Failed to convert index " + me.Environment["INDEX"] + " into a number!"
		return
	}
	SOURCEIDX = wimindex - 1
	me.Environment["NAME"] = SOURCEIMG.IMAGE[SOURCEIDX].NAME.Text
	me.Environment["DESC"] = SOURCEIMG.IMAGE[SOURCEIDX].DESCRIPTION.Text
	me.Environment["ROOT"] = SOURCEIMG.IMAGE[SOURCEIDX].WINDOWS.SYSTEMROOT.Text
	me.Environment["MAJOR"] = SOURCEIMG.IMAGE[SOURCEIDX].WINDOWS.VERSION.MAJOR.Text
	me.Environment["MINOR"] = SOURCEIMG.IMAGE[SOURCEIDX].WINDOWS.VERSION.MINOR.Text
	me.Environment["BUILD"] = SOURCEIMG.IMAGE[SOURCEIDX].WINDOWS.VERSION.BUILD.Text
	me.Environment["SPLVL"] = SOURCEIMG.IMAGE[SOURCEIDX].WINDOWS.VERSION.SPLEVEL.Text
	me.Environment["SPVER"] = SOURCEIMG.IMAGE[SOURCEIDX].WINDOWS.VERSION.SPBUILD.Text
	me.Environment["SUM"] = SOURCEIMG.IMAGE[SOURCEIDX].Summary

	switch me.Environment["MAJOR"] {
	case "6":
		switch me.Environment["MINOR"] {
		case "0":
			me.Environment["WINDOWS"] = "Vista"
			if me.Environment["BOOT"] == "MBR7" {
				me.Environment["BOOT"] = "MBRVISTA"
			}
		case "1":
			me.Environment["WINDOWS"] = "7"
		case "2":
			me.Environment["WINDOWS"] = "8"
		case "3":
			me.Environment["WINDOWS"] = "8.1"
		}
	case "10":
		switch me.Environment["MINOR"] {
		case "0":
			build, _ := strconv.Atoi(me.Environment["BUILD"])
			if build < 20000 {
				me.Environment["WINDOWS"] = "10"
			} else {
				me.Environment["WINDOWS"] = "11"
			}
		}
	}
	if windows, ok := me.Environment["WINDOWS"]; !ok || windows == "" {
		me.ErrorText("Unsupported NT version " + me.Environment["MAJOR"] + "." + me.Environment["MINOR"], "")
		return
	}

	guid, err := reghive.GenerateGuid()
	if err != nil {
		me.ErrorText("Failed to generate GUID for this install!", err.Error())
		return
	}
	me.Environment["GUID"] = guid

	if _, err := os.Stat("/sys/firmware/efi"); err != nil {
		me.Environment["BOOT"] = "MBR7"
	} else {
		me.Environment["BOOT"] = "EFI"
	}

	added := 0
	disks := osmgr.GetDisks()
	for i := 0; i < len(disks); i++ {
		disk := disks[i]
		added++
		menu.AddItem(fmt.Sprintf("%s (%s, /dev/%s)", disk.Model, disk.Size, disk.Block), "This will FORMAT your "+disk.Model+" and any data on it will be permanently destroyed!", "setvar DISK "+disk.Block, "menu install-windows-confirmation")
	}
	if added == 0 {
		menu.Title = "No disks were found!"
	}
	me.Redraw()
}

func hookInstallWindowsConfirmation(me *MenuEngine) {
	disks := osmgr.GetDisks()
	for i := 0; i < len(disks); i++ {
		disk := disks[i]
		if disk.Block != me.Environment["DISK"] {
			continue
		}
		me.Environment["DISKPRETTY"] = fmt.Sprintf("%s (%s, /dev/%s)", disk.Model, disk.Size, disk.Block)
		break
	}
	me.Redraw()
}

func hookInstallWindows(me *MenuEngine) {
	me.Redraw()
	me.Lock()
	defer me.Unlock()

	//Perform the Windows installation!
	wim := me.Environment["WIM"]
	index := me.Environment["INDEX"]
	disk := me.Environment["DISK"]
	target := "/dev/" + disk
	img := me.Environment["windowsimg"]
	boot := me.Environment["windowsboot"]

	menu := me.Menus["install-windows"]
	menu.Items = make([]*MenuItem, 0)
	me.Redraw()

	printf("* Unmounting %s\n", target)
	partitions, err := ioutil.ReadFile("/proc/partitions")
	if err != nil {
		me.ErrorText("Failed to read partition tables!", err.Error())
		return
	}
	matcher, _ := regexp.Compile("(?:\\d) ([a-z]+)(\\d)")
	matches := matcher.FindAllStringSubmatch(string(partitions), -1)
	if len(matches) == 0 {
		me.ErrorText("No disks were found!", "")
		return
	}
	for i := 0; i < len(matches); i++ {
		disk := "/dev/" + matches[i][1]
		if disk != target {
			//menu.AddItem("* Skipping disk "+disk+" because it is not "+target, "", "note", "")
			//me.Redraw()
			continue
		}
		part := matches[i][2]
		drive := disk + part
		//menu.AddItem(fmt.Sprintf("\n%v\n%v\n", []byte(target), []byte(drive)), "", "note", "")
		//menu.AddItem("* Unmounting drive "+drive, "", "note", "")
		//me.Redraw()
		_, _ = Run("umount", drive)
	}

	printf("* Making sure %s is unmounted properly\n", target)
	mounts, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		me.ErrorText("Failed to read mounts!", err.Error())
		return
	}
	matcher, _ = regexp.Compile("(/dev/[a-z]+)(\\d)")
	matches = matcher.FindAllStringSubmatch(string(mounts), -1)
	if len(matches) > 0 {
		for i := 0; i < len(matches); i++ {
			disk := matches[i][1]
			if disk != target {
				//menu.AddItem("* Skipping disk "+disk+" because it is not "+target, "", "note", "")
				//me.Redraw()
				continue
			}
			part := matches[i][2]
			drive := disk + part
			me.ErrorText("Failed to unmount " + drive + "!", "")
			return
		}
	}

	println("* Determining how to partition Windows")
	targetBOOT := target
	targetIMG := target
	targetREC := target
	if len(disk) > 6 && string(disk[:6]) == "mmcblk" {
		targetBOOT += "p"
		targetIMG += "p"
		targetREC += "p"
	}
	switch me.Environment["BOOT"] {
	case "MBRVISTA", "MBR7":
		targetBOOT += "1"
		targetIMG += "2"
	case "EFI":
		targetBOOT += "1"
		targetIMG += "3"
		targetREC += "2"
	}

	///*
	printf("* Creating Windows partitions on %s\n", target)
	switch me.Environment["BOOT"] {
	case "MBRVISTA", "MBR7":
		output, err := Run("sh", "fdisk_mbr.sh", target)
		if err != nil {
			me.ErrorText("Failed to call fdisk!", err.Error() + "\n\n" + string(output))
			return
		}
	case "EFI":
		output, err := Run("sh", "fdisk_gpt.sh", target)
		if err != nil {
			me.ErrorText("Failed to call fdisk!", err.Error() + "\n\n" + string(output))
			return
		}
	}

	switch me.Environment["BOOT"] {
	case "MBRVISTA", "MBR7":
		printf("* Formatting %s to NTFS\n", targetBOOT)
		output, err := Run("mkfs.ntfs", "--quick", "--label", "System Reserved", targetBOOT)
		if err != nil {
			me.ErrorText("Failed to format boot partition!", err.Error() + "\n\n" + string(output))
			return
		}
	case "EFI":
		printf("* Formatting %s to FAT32\n", targetBOOT)
		output, err := Run("mkfs.fat", "-F", "32", targetBOOT)
		if err != nil {
			me.ErrorText("Failed to format boot partition!", err.Error() + "\n\n" + string(output))
			return
		}
	}

	if targetREC != target {
		printf("* Formatting %s to NTFS\n", targetREC)
		output, err := Run("mkfs.ntfs", "--quick", "--label", "Recovery", targetREC)
		if err != nil {
			me.ErrorText("Failed to format recovery partition!", err.Error() + "\n\n" + string(output))
			return
		}
	}

	printf("* Formatting %s to NTFS\n", targetIMG)
	output, err := Run("mkfs.ntfs", "--quick", "--label", "Windows", targetIMG)
	if err != nil {
		me.ErrorText("Failed to format Windows partition!", err.Error() + "\n\n" + string(output))
		return
	}

	switch me.Environment["BOOT"] {
	case "MBRVISTA":
		printf("* Installing Windows Vista Master Boot Record (MBR) to %s\n", target)
		output, err := Run("ms-sys", "--mbrvista", target)
		if err != nil {
			me.ErrorText("Failed to install the Windows Vista MBR!", err.Error() + "\n\n" + string(output))
			return
		}
	case "MBR7":
		printf("* Installing Windows 7 Master Boot Record (MBR) to %s\n", target)
		output, err := Run("ms-sys", "--mbr7", target)
		if err != nil {
			me.ErrorText("Failed to install the Windows 7 MBR!", err.Error() + "\n\n" + string(output))
			return
		}
	}

	switch me.Environment["BOOT"] {
	case "MBRVISTA", "MBR7":
		printf("* Installing Windows 7 NTFS boot record to %s\n", targetBOOT)
		output, err := Run("ms-sys", "--ntfs", targetBOOT)
		if err != nil {
			me.ErrorText("Failed to install the Windows 7 NTFS boot record!", err.Error() + "\n\n" + string(output))
			return
		}
	}

	printf("* Applying %s:%s to %s\n", wim, index, targetIMG)
	err = RunRealtime("wimapply", wim, index, targetIMG)
	if err != nil {
		me.ErrorText("Failed to apply the Windows image!", err.Error())
		return
	}
	//*/

	printf("* Mounting Windows partitions from %s\n", target)
	output, err = Run("mount", "-o", "rw", targetBOOT, boot)
	//output, err := Run("mount", "-o", "rw", targetBOOT, boot)
	if err != nil {
		me.ErrorText("Failed to mount the boot partition!", err.Error() + "\n\n" + string(output))
		return
	}
	output, err = Run("mount", "-o", "ro", targetIMG, img)
	if err != nil {
		me.ErrorText("Failed to mount the Windows partition!", err.Error() + "\n\n" + string(output))
		return
	}

	printf("* Installing the Windows bootloader to %s\n", targetBOOT)

	println("* - PCAT")
	err = copy.Copy(img+"/Windows/Boot/PCAT", boot+"/Boot")
	if err != nil {
		me.ErrorText("Failed to install PCAT!", err.Error())
		return
	}
	err = os.Rename(boot+"/Boot/bootmgr", boot+"/bootmgr")
	if err != nil {
		me.ErrorText("Failed to install bootmgr!", err.Error())
		return
	}
	err = copy.Copy(img+"/Windows/Boot/Fonts", boot+"/Boot/Fonts")
	if err != nil {
		me.ErrorText("Failed to install Fonts!", err.Error())
		return
	}

	switch me.Environment["BOOT"] {
	case "EFI":
		switch me.Environment["WINDOWS"] {
		case "8", "8.1", "10", "11":
			err = copy.Copy(img+"/Windows/Boot/Resources", boot+"/Boot/Resources")
			if err != nil {
				me.ErrorText("Failed to install Resources!", err.Error())
				return
			}
			err = os.Rename(boot+"/Boot/bootnxt", boot+"/BOOTNXT")
			if err != nil {
				me.ErrorText("Failed to install bootnxt!", err.Error())
				return
			}
		}

		println("* - EFI")
		err = os.MkdirAll(boot+"/EFI/Boot", 0777)
		if err != nil {
			me.ErrorText("Failed to create EFI boot folders!", err.Error())
			return
		}
		err = os.MkdirAll(boot+"/EFI/Microsoft", 0777)
		if err != nil {
			me.ErrorText("Failed to create EFI boot folders!", err.Error())
			return
		}
		err = copy.Copy(img+"/Windows/Boot/EFI", boot+"/EFI/Microsoft/Boot")
		if err != nil {
			me.ErrorText("Failed to install EFI!", err.Error())
			return
		}
		err = copy.Copy(boot+"/EFI/Microsoft/Boot/bootmgfw.efi", boot+"/EFI/Boot/bootx64.efi")
		if err != nil {
			me.ErrorText("Failed to copy EFI boot manager!", err.Error())
			return
		}
		err = copy.Copy(boot+"/Boot/Fonts", boot+"/EFI/Microsoft/Boot/Fonts")
		if err != nil {
			me.ErrorText("Failed to copy EFI Fonts!", err.Error())
			return
		}

		switch me.Environment["WINDOWS"] {
		case "8", "8.1", "10", "11":
			err = copy.Copy(boot+"/Boot/Resources", boot+"/EFI/Microsoft/Boot/Resources")
			if err != nil {
				me.ErrorText("Failed to copy EFI Resources!", err.Error())
				return
			}
		}
	}

	println("* Adjusting BCD-Template hive for the new install")
	err = copy.Copy(img+"/Windows/System32/config/BCD-Template", boot+"/BCD")
	if err != nil {
		me.ErrorText("Failed to copy BCD-Template into staging area!", err.Error())
		return
	}

	bcd, err := reghive.OpenRegistryHive(boot + "/BCD")
	if err != nil {
		me.ErrorText("Failed to open BCD!", err.Error())
		return
	}

	rootNode, err := bcd.GetNode("/")
	if err != nil {
		me.ErrorText("Failed to get root node from BCD!", err.Error())
		return
	}

	println("* Finished adjusting BCD!")
	printf("* - Root node: %s\n", rootNode.Name)
	//println(rootNode.String())

	println("* - Injecting BCD into PCAT")
	err = copy.Copy(boot+"/BCD", boot+"/Boot/BCD")
	if err != nil {
		me.ErrorText("Failed to install BCD into PCAT!", err.Error())
		return
	}

	switch me.Environment["BOOT"] {
	case "EFI":
		println("* - Injecting BCD into EFI")
		err = copy.Copy(img+"/Windows/System32/config/BCD-Template", boot+"/EFI/Microsoft/Boot/BCD")
		if err != nil {
			me.ErrorText("Failed to install BCD into EFI!", err.Error())
			return
		}
	}

	println("* Unmounting Windows partitions")
	output, err = Run("umount", boot)
	if err != nil {
		me.ErrorText("Failed to unmount boot!", err.Error() + "\n\n" + string(output))
		return
	}
	output, err = Run("umount", img)
	if err != nil {
		me.ErrorText("Failed to unmount Windows!", err.Error() + "\n\n" + string(output))
		return
	}

	me.Unlock()
	printf("\nDone!\n")
	time.Sleep(1 * time.Second)

	menu.AddItem("Reboot to complete the install", "Make sure to remove this installation medium before pressing enter!", "internal", "reboot")
	menu.AddItem("Shut down to complete the install later", "Make sure to remove this installation medium before pressing enter!", "internal", "shutdown")
	menu.AddItem("", "", "divider", "2")
	menu.AddItem("Return to setup to do something else", "Takes you back to the install and recovery choices", "menu", "setup")
	menu.NoSelector = false
	me.ItemCursor = len(menu.Items) - 4
	me.Redraw()
}

