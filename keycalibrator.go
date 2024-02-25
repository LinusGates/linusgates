package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/JoshuaDoes/json"
)

var (
	ERR_CANCELLED = fmt.Errorf("calibrator: cancelled")
)

type MenuKeycodeBinding struct {
	Keycode   uint16 `json:"keycode"`
	Action    string `json:"action"`
	OnRelease bool   `json:"onRelease"`
}

func bindKeys() {
	for keyboard, bindings := range keyCalibration {
		kl, err := NewKeycodeListener(keyboard)
		if err != nil {
			panic(fmt.Sprintf("error listening to keyboard %s: %v", keyboard, err))
		}
		for _, binding := range bindings {
			var action func()
			switch binding.Action {
			case "prevItem":
				action = menuEngine.PrevItem
			case "nextItem":
				action = menuEngine.NextItem
			case "selectItem":
				action = menuEngine.Action
			default:
				panic("unknown action: " + binding.Action)
			}
			kl.Bind(binding.Keycode, binding.OnRelease, action)
		}
		go kl.Run()
	}
}

type KeyCalibration struct {
	Ready  bool
	Cancel bool
	Action string
	KLs    []*KeycodeListener
}

func (kc *KeyCalibration) Input(keyboard string, keycode uint16, onRelease bool) {
	if kc.Cancel {
		return
	}
	if !kc.Ready {
		kc.Cancel = true
		return
	}
	if kc.Action == "" || kc.Action == "cancel" {
		kc.Action = ""
		return
	}
	if onRelease {
		return
	}
	if keyCalibration[keyboard] == nil {
		keyCalibration[keyboard] = make([]*MenuKeycodeBinding, 0)
	}
	keyCalibration[keyboard] = append(keyCalibration[keyboard], &MenuKeycodeBinding{
		Keycode:   keycode,
		Action:    kc.Action,
		OnRelease: true,
	})
	kc.Action = ""
}

func calibrate() error {
	//Generate a key calibration file if one doesn't exist yet
	calibrator := &KeyCalibration{KLs: make([]*KeycodeListener, 0)}

	//Get a list of keyboards
	keyboards := make([]string, 0)
	err := filepath.Walk("/dev/input", func(path string, info os.FileInfo, err error) error {
		if len(path) < 16 || string(path[:16]) != "/dev/input/event" {
			return nil
		}
		keyboards = append(keyboards, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking inputs: %v", err)
	}

	//Bind all keyboards to calibrator input
	for _, keyboard := range keyboards {
		kl, err := NewKeycodeListener(keyboard)
		if err != nil {
			return fmt.Errorf("error listening to walked keyboard %s: %v", keyboard, err)
		}
		kl.RootBind = calibrator.Input
		calibrator.KLs = append(calibrator.KLs, kl)
		go kl.Run()
	}

	//Start calibrating!
	stages := 6
	for stage := 0; stage < stages; stage++ {
		switch stage {
		case 0:
			keyCalibrationJSON, err := ioutil.ReadFile(keyCalibrationFile)
			if err == nil {
				keyCalibration = make(map[string][]*MenuKeycodeBinding)
				err = json.Unmarshal(keyCalibrationJSON, &keyCalibration)
				if err != nil {
					stage = 1
					continue
				}

				clear()
				println("Press any key within\n5 seconds to recalibrate.\n")
				calibrator.Ready = true
				calibrator.Action = "cancel"
				timeout := time.Now()
				for calibrator.Action != "" {
					if time.Now().Sub(timeout).Seconds() > 5 {
						break
					}
					time.Sleep(time.Millisecond * 100)
				}
				if time.Now().Sub(timeout).Seconds() < 5 {
					calibrator.Action = ""
					println("Recalibration time!")
					time.Sleep(time.Second * 2)
					continue
				}
				stage = stages-1 //Skip to the end of the stages
			}
		case 1:
			calibrator.Ready = false
			keyCalibration = make(map[string][]*MenuKeycodeBinding)
			clear()
			println("Welcome to the calibrator!\n")
			println("Press any key to cancel.\n")
			time.Sleep(time.Second * 2)
			if calibrator.Cancel { return ERR_CANCELLED }
			println("Controllers and remotes\nare also supported.\n")
			time.Sleep(time.Second * 2)
			if calibrator.Cancel { return ERR_CANCELLED }
			println("This is a guided process.\n")
			time.Sleep(time.Second * 2)
			if calibrator.Cancel { return ERR_CANCELLED }
			println("Get ready!\n")
			if calibrator.Cancel { return ERR_CANCELLED }
			time.Sleep(time.Second * 3)
			if calibrator.Cancel { return ERR_CANCELLED }
		case 2:
			clear()
			calibrator.Ready = true
			calibrator.Action = "nextItem"
			printf("\n")
			println("Press any key to use to\nnavigate down in a menu.\n")
			println("Recommended: volume down")
			for calibrator.Action != "" {
			}
		case 3:
			calibrator.Action = "prevItem"
			printf("\n")
			println("Press any key to use to\nnavigate up in a menu.\n")
			println("Recommended: volume up")
			for calibrator.Action != "" {
			}
		case 4:
			calibrator.Action = "selectItem"
			printf("\n")
			println("Press any key to use to\nselect a menu item.\n")
			println("Recommended: touch screen")
			for calibrator.Action != "" {
			}
		case 5:
			clear()
			println("Saving results...\n")
			keyboards, err := json.Marshal(keyCalibration, true)
			if err != nil {
				return fmt.Errorf("error encoding calibration results: %v", err)
			}
			keyboardsFile, err := os.Create(keyCalibrationFile)
			if err != nil {
				return fmt.Errorf("error creating calibration file: %v", err)
			}
			defer keyboardsFile.Close()
			_, err = keyboardsFile.Write(keyboards)
			if err != nil {
				return fmt.Errorf("error writing calibration file: %v", err)
			}
			//println(string(keyboards))
			//println("Calibration complete!")
			//time.Sleep(time.Second * 2)
			//calibrator.Ready = false
		}
	}

	for i := 0; i < len(calibrator.KLs); i++ {
		calibrator.KLs[i].Close()
	}
	return nil
}
