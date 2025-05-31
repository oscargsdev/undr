package com.undr.demo.controller;

import com.undr.demo.domain.MusicGenre;
import com.undr.demo.repository.MusicGenreRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class PrototypeController {
    @Autowired
    private MusicGenreRepository musicGenreRepository;

    @GetMapping("/musicgenres")
    public ResponseEntity<Iterable<MusicGenre>> getMusicGenres(){
        return ResponseEntity.ok(musicGenreRepository.findAll().stream().toList());
    }

}
