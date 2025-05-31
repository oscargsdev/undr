package com.undr.demo.dto;

import lombok.Getter;
import lombok.Setter;

import java.util.List;

@Getter
@Setter
public class ProjectGenreDTO {
    private MusicGenreCreationDTO mainGenre;
    private List<MusicGenreCreationDTO> subGenres;
}
