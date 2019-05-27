# Ubuntu Core 20 Experimental Tooling

This repository contains early experimental tooling for working with concepts
of Ubuntu Core 20. Core 20 will ship around April 2020 and will feature a new
recovery system and full disk encryption.

## What to expect here?

This repository contains extremely early prototype for a piece of code that
runs in the initial ram disk (initramfs), responsible for performing boot mode
selection and setting up initial mount namespace.

At the heart of the recovery is a new partitioning scheme. The existing
"writable" partition is accompanied by a new, yet officially unnamed, "systems"
partition. The systems partition contains *sets* of (base, kernel, gadget,
snapd) snaps, any of which can be selected at boot using the platform native
bootloader.

The system will differentiate **RUN** mode, which is similar to what Ubuntu
Core systems did since the 16 release from **RECOVERY** modes. In the RUN mode
a most-recent set of snaps is used and the system boots normally, unlocking
access to the writable partition and the data stored therein.

In the RECOVERY mode access to the writable partition is retained but the
partition can be also wiped and re-created, from the set of snaps stored in the
systems partition.

This design can be expanded to create additional sub-modes of RECOVERY, where
certain data is retained while other data is reset, but the general idea
stands.

## Hacking.

To work with this repository you will need `snap`, `make`, `uuidgen`, `sdisk`,
`grub2-editenv`, `qemu-system-x86_64` as well as the `ovmf` UEFI firmware for
qemu. The `snap` command is only used to download the `pc-kernel` snap. Having
done this (the results can be safely cached) builds are entirely offline.
Reading the `Makefile` is recommended as it explains many things in greater
detail.

The build system is highly parallel and incremental, being able to build a
bootable hard disk image under ten seconds. Subsequent changes to files in the
repository trigger incremental rebuilds, supporting very fast, offline
iteration.

There is no support for testing yet, though a rudimentary set of tests, both
unit and integration, is planned soon.

## Execution flow

The experiments stages a disk image consisting of EFI-only (legacy mode is not
supported yet) GRUB boot loader, using an embedded configuration file to run a
kernel and initrd coming from the pc-kernel snap. The initramfs is augmented
with an all new snapd-initrd-boot executable, that drives the early boot
process.

In the experiment it spawns shell, in the final design it will pass on to the
real init system.
