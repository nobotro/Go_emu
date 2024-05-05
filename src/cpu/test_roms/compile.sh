
rm *.o
rm *.out
rm rom
riscv32-unknown-elf-gcc -O0  starter.s PrintDigits_source.c -Ttext 0 -ffreestanding -fno-stack-protector  -fno-pie -march=rv32i -mabi=ilp32 -c
riscv32-unknown-elf-ld starter.o PrintDigits_source.o -T link.ld
riscv32-unknown-elf-objcopy -O binary a.out PrintDigits_rom
