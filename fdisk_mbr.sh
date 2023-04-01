#!/bin/bash
if [ -z "$1" ]; then
	echo "** MISSING DEVICE! **"
	exit 1
fi

## https://superuser.com/a/984637
# to create the partitions programatically (rather than manually)
# we're going to simulate the manual input to fdisk
# The sed script strips off all the comments so that we can
# document what we're doing in-line with the actual commands
# Note that a blank line (commented as "defualt" will send a empty
# line terminated with a newline to take the fdisk default.
sed -e 's/\s*\([\+0-9a-zA-Z]*\).*/\1/' << EOF | fdisk "$1"
  o # overwrite the partition table with MBR
  n # new partition
  p # primary
  1 # Windows bootloader
    # default - start at beginning of disk
  +128M # 128 MB boot parttion
  n # new partition
  p # primary
  2 # Windows
    # default - start after last partition
    # default - fill to end of disk
  t # change partition type
  1 # change partition 1
  7 # HPFS/NTFS/exFAT
  t # change partition type one more time
  2 # change partition 3
  7 # HPFS/NTFS/exFAT
  a # make a partition bootable
  1 # bootable partition is partition 1 -- /dev/sdc1
  p # print the in-memory partition table
  w # write the partition table
  q # and we're done
EOF
sync
sleep 3
