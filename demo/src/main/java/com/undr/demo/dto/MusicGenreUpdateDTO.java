package com.undr.demo.dto;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
@AllArgsConstructor
public class MusicGenreUpdateDTO {
    Long musicGenreId;
    String musicGenreName;
}
