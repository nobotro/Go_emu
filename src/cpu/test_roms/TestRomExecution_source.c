#include <stdint.h>

volatile uint32_t* SCREEN_ADDRESS = (volatile uint32_t*)30032;
volatile uint32_t* count = (volatile uint32_t*)30028;
void main(){

    int a=99;
    int g=22;

    int counter=0;
    while(1)
    {
      for(int i=0;i<4;i++)
        *(SCREEN_ADDRESS+i)=a+g-1;

      if (counter<=1000){
        *(count)=counter;
        counter++;
      }
    }
}




riscv32-unknown-elf-objdump -d a.out

a.out:     file format elf32-littleriscv


Disassembly of section .text:

00000000 <_start>:
   0:	00007137          	lui	sp,0x7
   4:	53010113          	addi	sp,sp,1328 # 7530 <count+0x7490>
   8:	004000ef          	jal	ra,c <main>

0000000c <main>:
   c:	fe010113          	addi	sp,sp,-32
  10:	00812e23          	sw	s0,28(sp)
  14:	02010413          	addi	s0,sp,32
  18:	06300793          	li	a5,99
  1c:	fef42223          	sw	a5,-28(s0)
  20:	01600793          	li	a5,22
  24:	fef42023          	sw	a5,-32(s0)
  28:	fe042623          	sw	zero,-20(s0)
  2c:	fe042423          	sw	zero,-24(s0)
  30:	0380006f          	j	68 <main+0x5c>
  34:	fe442703          	lw	a4,-28(s0)
  38:	fe042783          	lw	a5,-32(s0)
  3c:	00f707b3          	add	a5,a4,a5
  40:	fff78693          	addi	a3,a5,-1
  44:	09c02703          	lw	a4,156(zero) # 9c <SCREEN_ADDRESS>
  48:	fe842783          	lw	a5,-24(s0)
  4c:	00279793          	slli	a5,a5,0x2
  50:	00f707b3          	add	a5,a4,a5
  54:	00068713          	mv	a4,a3
  58:	00e7a023          	sw	a4,0(a5)
  5c:	fe842783          	lw	a5,-24(s0)
  60:	00178793          	addi	a5,a5,1
  64:	fef42423          	sw	a5,-24(s0)
  68:	fe842703          	lw	a4,-24(s0)
  6c:	00300793          	li	a5,3
  70:	fce7d2e3          	bge	a5,a4,34 <main+0x28>
  74:	fec42703          	lw	a4,-20(s0)
  78:	3e800793          	li	a5,1000
  7c:	fae7c8e3          	blt	a5,a4,2c <main+0x20>
  80:	0a002783          	lw	a5,160(zero) # a0 <count>
  84:	fec42703          	lw	a4,-20(s0)
  88:	00e7a023          	sw	a4,0(a5)
  8c:	fec42783          	lw	a5,-20(s0)
  90:	00178793          	addi	a5,a5,1
  94:	fef42623          	sw	a5,-20(s0)
  98:	f95ff06f          	j	2c <main+0x20>
