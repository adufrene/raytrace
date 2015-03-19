camera {
  location  <0, 0, 14>
  up        <0,  1,  0>
  right     <1.33333, 0,  0>
  look_at   <0, 0, 0>
}

light_source {<-100, 100, 100> color rgb <1.0, 1.0, 1.0>}

sphere { <0, 2, -9>, 5
  pigment { color rgbf <0.8, 0.0, 1.0, 0.0>}
  finish {ambient 0.2 diffuse 0.4 specular 0.9 reflection 1.0 roughness 0.02 }
}              

sphere { <1, 1, 1.5>, 2
  pigment { color rgbf <0.0, 0.0, 0.0 0.9>}
  finish {ambient 0.1 diffuse 0.1 specular 0.3 roughness 0.001 reflection 0.3 refraction 1.0 ior 1.33}
}  

plane { <0,1,0> , -4
  pigment {color rgb <0.4, 0.4, 0.7>}
  finish {ambient 0.4 diffuse 0.2 reflection 1.0}
}
                               
                               
                               
