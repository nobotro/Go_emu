


MEMORY
   {
     ROM (xrw)       : ORIGIN = 0, LENGTH = 15000
     STACK (xrw)       : ORIGIN = 30000, LENGTH = 15000
   }


SECTIONS
 {
   
   
   .text : { *(.text) } 
   .data : { *(.data) }
   .bss  : { *(.bss)  } > ROM
   .stack : { *(.stack) }

 }
