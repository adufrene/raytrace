camera {
    location <0, 0, 14>
    up <0, 1, 0>
    right <1.33333, 0, 0>
    look_at <0, 0, 0>
}

light_source {<-100, 100, 100> color rgb <1.5, 1.5, 1.5>}

sphere { <5.4, -2, -2>, 4
    pigment {color rgb <0.8, 0.8, 0.7>}
    finish {ambient 0.2 diffuse 0.4 reflection 0.9}
}

sphere { <0, 4, 4>, 2
    pigment {color rgb <1.0, 0.1, 0.1>}
    finish {reflection 0.2}
}

sphere { <-5.4, -2, -2>, 4
    pigment {color rgb <0.1, 0.1, 1.0>}
    finish {ambient 0.2 diffuse 0.4 specular 0.6 roughness 0.01 reflection 0.8}
}

sphere { <0, 0, 20>, 5
    pigment {color rgb <0.1, 1.0, 0.1>}
    finish {ambient 0.1 diffuse 0.2 reflection 0.8}
}
