package com.undr.demo.service;

import com.undr.demo.domain.MusicGenre;
import com.undr.demo.dto.MusicGenreCreationDTO;
import com.undr.demo.dto.MusicGenreUpdateDTO;

public interface MusicGenreService {
    MusicGenre createMusicGenre(MusicGenreCreationDTO musicGenreData);
    MusicGenre updateMusicGenre(MusicGenreUpdateDTO musicGenreData);
    MusicGenre getMusicGenreById(long musicGenreId);
    void deleteMusicGenre(long musicGenreId);
    Iterable<MusicGenre> getAllMusicGenres();
}
