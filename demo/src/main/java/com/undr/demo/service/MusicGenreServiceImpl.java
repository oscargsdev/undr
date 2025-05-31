package com.undr.demo.service;

import com.undr.demo.domain.MusicGenre;
import com.undr.demo.dto.MusicGenreCreationDTO;
import com.undr.demo.dto.MusicGenreUpdateDTO;
import com.undr.demo.repository.MusicGenreRepository;
import org.modelmapper.ModelMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class MusicGenreServiceImpl implements MusicGenreService {
    @Autowired
    private final MusicGenreRepository musicGenreRepository;

    public MusicGenreServiceImpl(MusicGenreRepository musicGenreRepository) {
        this.musicGenreRepository = musicGenreRepository;
    }

    private final ModelMapper mapper = new ModelMapper();

    @Override
    public MusicGenre createMusicGenre(MusicGenreCreationDTO musicGenreData) {
        MusicGenre newMusicGenre = mapper.map(musicGenreData, MusicGenre.class);
        musicGenreRepository.save(newMusicGenre);
        return newMusicGenre;
    }

    @Override
    public MusicGenre updateMusicGenre(MusicGenreUpdateDTO musicGenreData) {
        MusicGenre updatedMusicGenre = musicGenreRepository.findById(musicGenreData.getMusicGenreId()).get();

        mapper.map(musicGenreData, updatedMusicGenre);
        musicGenreRepository.save(updatedMusicGenre);

        return updatedMusicGenre;
    }

    @Override
    public MusicGenre getMusicGenreById(long musicGenreId) {
        MusicGenre musicGenre = musicGenreRepository.findById(musicGenreId).get();

        return musicGenre;
    }

    @Override
    public void deleteMusicGenre(long musicGenreId) {
        MusicGenre deletedMusicGenre = musicGenreRepository.findById(musicGenreId).get();

        musicGenreRepository.delete(deletedMusicGenre);
    }

    @Override
    public Iterable<MusicGenre> getAllMusicGenres() {
        return musicGenreRepository.findAll().stream().toList();
    }
}
