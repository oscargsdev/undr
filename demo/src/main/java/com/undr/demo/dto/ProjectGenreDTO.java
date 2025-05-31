package com.undr.demo.dto;

import lombok.Getter;
import lombok.Setter;

import java.util.List;

@Getter
@Setter
public class ProjectGenreDTO{
    private MusicGenreDTO mainGenre;
    private List<MusicGenreDTO> subGenres;
}
