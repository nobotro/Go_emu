#include <stdint.h>

volatile uint32_t* SCREEN_ADDRESS = (volatile uint32_t*)30032;

void draw_pixel(uint32_t x,uint32_t y,uint32_t pixel ){

    *(SCREEN_ADDRESS+(y*64) +x) = pixel;
}
uint32_t font8x8_basic[10][8]= {

    {0x3E, 0x63, 0x73, 0x7B, 0x6F, 0x67, 0x3E, 0x00},   // U+0030 (0)
    {0x0C, 0x0E, 0x0C, 0x0C, 0x0C, 0x0C, 0x3F, 0x00},   // U+0031 (1)
    {0x1E, 0x33, 0x30, 0x1C, 0x06, 0x33, 0x3F, 0x00},   // U+0032 (2)
    {0x1E, 0x33, 0x30, 0x1C, 0x30, 0x33, 0x1E, 0x00},   // U+0033 (3)
    {0x38, 0x3C, 0x36, 0x33, 0x7F, 0x30, 0x78, 0x00},   // U+0034 (4)
    {0x3F, 0x03, 0x1F, 0x30, 0x30, 0x33, 0x1E, 0x00},   // U+0035 (5)
    {0x1C, 0x06, 0x03, 0x1F, 0x33, 0x33, 0x1E, 0x00},   // U+0036 (6)
    {0x3F, 0x33, 0x30, 0x18, 0x0C, 0x0C, 0x0C, 0x00},   // U+0037 (7)
    {0x1E, 0x33, 0x33, 0x1E, 0x33, 0x33, 0x1E, 0x00},   // U+0038 (8)
    {0x1E, 0x33, 0x33, 0x3E, 0x30, 0x18, 0x0E, 0x00}    // U+0039 (9)
};


void main(){
  uint32_t x = 0; //cursor x
  uint32_t y = 0;  //cursor y
  uint32_t c=0;
  while(c<10){
      if(x<54) x+=9;
      else{
          if(y<22){
            y+=9;
            x=0;
          }
          else return;
      }
      for(uint32_t i =0;i<8;i++)
        {
          for(uint32_t j =0;j<8;j++)
          {

              draw_pixel(x+j,y+i,(font8x8_basic[c][i]>>j)&1);

          }
        }
     c++;
  }
  while(1){

  }
}
