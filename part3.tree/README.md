# Partition 3 (Unencrypted Assets)

This partition holds the kernel, wrapped inside a squashfs, that we want to
boot. It will also contain the boot base, the gadget, the snapd snap and
(perhaps) other files, as required.

## Layout

The filesystem tree in the SYSTEMS partition looks like this:

`/systems/ - root of the hierarchy containing (kernel, base, gadget, snapd) sets.
`/systems/<set-name>/ - a particular set
`/systems/<set-name>/boot/grubenv - a GRUB environment file defining "snapd\_kernel"
`/snaps/ - spool directory for snaps and assertions (currently unused)

## Experiment behavior

This only contains the `pc-kernel` snap from the "18" channel. It is downloaded automatically by the makefile.
