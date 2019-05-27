# Partition 2 (EFI System Boot)

This partition holds the boot-loader for EFI systems. It really only contains
the "EFI" directory and everything inside.

## Layout

- `EFI/boot/bootx64.efi` - EFI boot-loader file (grubx64 from `pc` gadget)
- `EFI/ubuntu/grub.cfg` - Hand-crafted GRUB configuration file
