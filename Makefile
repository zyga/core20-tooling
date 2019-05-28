#!/usr/bin/make -f
.PHONY: all
all: disk.img
.PHONY: clean

MiB = 1048576
GiB = 1073741824

# Both dd and sdisk require block count and block size instead of a plain size.
# In this script we use sectors / blocks of 512 bytes.
sector_size = 512
# The entire image should be this:
disk_bytes_size = $(shell echo $$(( 2 * $(GiB) )))

# The first partition has no formatted filesystem and is designed to hold the
# 1st stage bootloader used in non-EFI boot. For the purpose of the demo those
# are not populated since we want to exercise the EFI boot path. This partition
# starts at an offset to take account for GPT and probably some other stuff I
# don't remember. The remaining partitions have their offsets calculated based
# on linear progression in the image.
part1_bytes_start   := $(shell echo $$(( 1 * $(MiB) )))
part1_bytes_size    := $(shell echo $$(( 1 * $(MiB) )))
part1_sectors_start := $(shell echo $$(( $(part1_bytes_start) / $(sector_size) )))
part1_sectors_size  := $(shell echo $$(( $(part1_bytes_size)  / $(sector_size) )))
# TODO: install non-EFI grub assets.
part1.img: Makefile
	truncate --size=0 $@
	truncate --size=$(part1_bytes_size) $@
clean::
	rm -f part1.img

# The second partition is the EFI System partition and contains boot loader
# executed by the firmware. It is a typical FAT filesystem with appropriate
# structure. You can most likely compare files here with what is mounted in
# /boot/efi on your EFI-booted Linux system.
part2_bytes_start   := $(shell echo $$(( $(part1_bytes_start) + $(part1_bytes_size) )))
part2_bytes_size    := $(shell echo $$(( 100 * $(MiB) )))
part2_sectors_start := $(shell echo $$(( $(part2_bytes_start) / $(sector_size) )))
part2_sectors_size  := $(shell echo $$(( $(part2_bytes_size)  / $(sector_size) )))

part2.img: PATH:=$(PATH):/usr/sbin
part2.img: Makefile $(shell find part2.tree -print)
	truncate --size=0 $@
	truncate --size=$(part2_bytes_size) $@
	mkfs.vfat --invariant -S $(sector_size) -v $@
	$(foreach fname,$(shell find part2.tree -mindepth 1 -maxdepth 1),mcopy -s -i $@ $(fname) ::;)

clean::
	rm -f part2.img

# The third partition is a core-specific, non-encrypted typical Linux
# filesystem. It contains recovery system as well as other files that are
# required in the INSTALL mode (where we want to recreate the what used to be
# called "writable" partition). This partition is larger to hold enough
# kernels, gadgets and core and snapd snaps to boot the system.
#
# This partition must, at least, hold enough snaps to perform full system
# re-installation. Even if in the RUN mode we could decrypt writable and get to
# some files there, we must have them here for recovery.
part3_bytes_start   := $(shell echo $$(( $(part2_bytes_start) + $(part2_bytes_size) )))
part3_bytes_size    := $(shell echo $$(( 1 * $(GiB) )))
part3_sectors_start := $(shell echo $$(( $(part3_bytes_start) / $(sector_size) )))
part3_sectors_size  := $(shell echo $$(( $(part3_bytes_size)  / $(sector_size) )))

part3.tree/snaps/ part3.tree/systems/factory/boot/:
	mkdir -p $@

# Download the kernel snap and mark a stamp file.
part3.tree/snaps/.stamp: | part3.tree/snaps/
	( cd $| && snap download pc-kernel --channel=18 )
	touch $@
part3.tree/systems/factory/boot/grubenv: part3.tree/snaps/.stamp | part3.tree/systems/factory/boot/
	snapd_kernel="$$(find part3.tree/snaps/ -name 'pc-kernel_*.snap' -printf '%f\n' 2>/dev/null | sort -n | head -n 1)"
	test -n "$$snapd_kernel"
	grub2-editenv $@ create && grub2-editenv $@ set snapd_kernel="$$snapd_kernel"
clean::
	rm -rf part3.tree/systems/
	rm -rf part3.tree/snaps/

# Build snapd-init.c into a init program in snapd.initrd directory
initrd.snapd/:
	mkdir -p $@
initrd.snapd/init: snapd-init.c | initrd.snapd/
	$(CC) -static -Wall -o $@ $< && strip $@
clean::
	rm -f initrd.snapd/init
fmt::
	clang-format -i $(wildcard *.c)

# Package snapd.initrd into a snapd-initrd.img in the systems partition
part3.tree/snapd-initrd.img: initrd.snapd/init
	find initrd.snapd/ \! -type d -printf '%P\n' | sort | cpio --create --directory=initrd.snapd --io-size=512 --format=newc --owner=0.0 | gzip > $@
clean::
	rm -f part3.tree/snapd-initrd.img

part3.img: PATH:=$(PATH):/usr/sbin
part3.img: Makefile part3.tree/systems/factory/boot/grubenv part3.tree/snapd-initrd.img $(shell find part3.tree -print)
	truncate --size=0 $@
	truncate --size=$(part3_bytes_size) $@
	fakeroot mkfs.ext4 \
		-q \
		-t ext4 \
		-U clear \
		-L snapd-systems \
		-d part3.tree \
		-O has_journal,extent,huge_file,flex_bg,metadata_csum,64bit,dir_nlink,extra_isize \
		$@
clean::
	rm -f part3.img

# The fourth partition is a core-specific, encrypted typical Linux filesystem.
# In this experiment it is NOT encrypted. It hosts all snap data as well as
# miscellaneous system writable files. In the core16 / core20 world this was
# the "writable" partition. It spans the rest of the disk.
part4_bytes_start   := $(shell echo $$(( $(part3_bytes_start) + $(part3_bytes_size) )))
# XXX: why do we need 16986?
part4_bytes_size    := $(shell echo $$(( $(disk_bytes_size) - $(part4_bytes_start) - 16986 )))
part4_sectors_start := $(shell echo $$(( $(part4_bytes_start) / $(sector_size) )))
part4_sectors_size  := $(shell echo $$(( $(part4_bytes_size)  / $(sector_size) )))

part4.img: PATH:=$(PATH):/usr/sbin
part4.img: Makefile $(shell find part4.tree -print)
	truncate --size=0 $@
	truncate --size=$(part4_bytes_size) $@
	fakeroot mkfs.ext4 \
		-q \
		-t ext4 \
		-U clear \
		-L snapd-data \
		-d part4.tree \
		-O has_journal,extent,huge_file,flex_bg,metadata_csum,64bit,dir_nlink,extra_isize \
		$@
clean::
	rm -f part4.img

# Our image will have the following composition:
#
# partition scheme: GTP
# partition 1: 1MB (raw, boot loader assets)
# partition 2: 100MB (fat, EFI system partition)
# partition 3: 1GB (ext4, ubuntu-core plaintext assets)
# partition 4: ~1GB (ext4, ubuntu-core encrypted assets)
#
# In the prototype phase partition 4 is not encrypted. It is just separate for
# design purposes so that we can keep the layout for future encryption work.
core_ns_uuid=$(shell uuidgen --sha1 --namespace @url --name https://ubuntu.com/core/)
.ONESHELL: disk.sfdisk
disk.sfdisk: Makefile
	cat <<SCRIPT >$@
	label: gpt
	unit: sectors
	start=$(part1_sectors_start) size=$(part1_sectors_size) type=21686148-6449-6E6F-744E-656564454649 bootable name="BIOS Boot Partition"
	start=$(part2_sectors_start) size=$(part2_sectors_size) type=C12A7328-F81F-11D2-BA4B-00A0C93EC93B name="EFI System Partition"
	start=$(part3_sectors_start) size=$(part3_sectors_size) type=$(shell uuidgen --sha1 --namespace $(core_ns_uuid) --name snapd-systems) name="Snapd Systems Partition"
	start=$(part4_sectors_start) size=$(part4_sectors_size) type=$(shell uuidgen --sha1 --namespace $(core_ns_uuid) --name snapd-data) name="Snapd Data Partition"
	SCRIPT
clean::
	rm -f disk.sfdisk

# We can update partitions individually and write them concurrently for more
# parallelism.
disk.img:: PATH:=$(PATH):/sbin:/usr/sbin # for sfdisk
disk.img:: disk.sfdisk
	truncate --size=0 $@
	truncate --size=$(disk_bytes_size) $@
	sfdisk --quiet $@ <$<
disk.img:: part1.img disk.sfdisk
	dd of=$@ seek=$(part1_sectors_start) count=$(part1_sectors_size) bs=$(sector_size) conv=notrunc,sparse if=$< status=none
disk.img:: part2.img disk.sfdisk
	dd of=$@ seek=$(part2_sectors_start) count=$(part2_sectors_size) bs=$(sector_size) conv=notrunc,sparse if=$< status=none
disk.img:: part3.img disk.sfdisk
	dd of=$@ seek=$(part3_sectors_start) count=$(part3_sectors_size) bs=$(sector_size) conv=notrunc,sparse if=$< status=none
disk.img:: part4.img disk.sfdisk
	dd of=$@ seek=$(part4_sectors_start) count=$(part4_sectors_size) bs=$(sector_size) conv=notrunc,sparse if=$< status=none
clean::
	rm -f disk.img

bios.flash: /usr/share/qemu/ovmf-x86_64.bin
	cp $< $@
clean::
	rm -f bios.flash

# Allow running the system via qemu + EFI.
#
# The options were chosen so that specific things happen:
#  -m 256    => reasonable "minimum memory we are likely to support"
#  -smp 4	 => expose to ordering issues and multi-threading
#  -snapshot => runs without clobbering the disk
# TODO: add virtual random number generator
# TODO: add virtual serial port for debugging
.PHONY: run
run: disk.img bios.flash
	qemu-system-x86_64 \
		-machine q35 \
		-monitor stdio \
		-m 256 \
		-smp 4 \
		-k en \
		-drive if=pflash,format=raw,file=bios.flash \
		-global isa-debugcon.iobase=0x402 -debugcon file:bios.log \
		-blockdev driver=file,node-name=disk_file,filename=$< \
		-blockdev driver=raw,node-name=disk,file=disk_file \
		-device virtio-blk-pci,drive=disk,bootindex=0 \
		-netdev id=net0,type=user \
		-device virtio-net-pci,netdev=net0,romfile=
clean::
	rm -f bios.log
