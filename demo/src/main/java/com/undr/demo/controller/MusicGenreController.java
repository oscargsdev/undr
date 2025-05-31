package com.undr.demo.controller;

import com.undr.demo.domain.MusicGenre;
import com.undr.demo.dto.MusicGenreCreationDTO;
import com.undr.demo.dto.MusicGenreUpdateDTO;
import com.undr.demo.service.MusicGenreService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.support.ServletUriComponentsBuilder;

import java.net.URI;

@RestController
@RequestMapping("/v1/musicgenres")
public class MusicGenreController {
    @Autowired
    private final MusicGenreService musicGenreService;

    public MusicGenreController(MusicGenreService musicGenreService) {
        this.musicGenreService = musicGenreService;
    }

    @GetMapping
    public ResponseEntity<Iterable<MusicGenre>> getAllMusicGenres() {
        return ResponseEntity.ok(musicGenreService.getAllMusicGenres());
    }

    @GetMapping("/{musicGenreId}")
    public ResponseEntity<MusicGenre> getMusicGenreById(@PathVariable long musicGenreId) {
        MusicGenre musicGenre = musicGenreService.getMusicGenreById(musicGenreId);

        return ResponseEntity.ok(musicGenre);
    }

    @PostMapping
    public ResponseEntity<MusicGenre> createMusicGenre(@RequestBody MusicGenreCreationDTO musicGenreData) {
        MusicGenre newMusicGenre = musicGenreService.createMusicGenre(musicGenreData);
        URI location = ServletUriComponentsBuilder.fromCurrentRequest().path("/{musicGenreId}").buildAndExpand(newMusicGenre.getMusicGenreId()).toUri();

        return ResponseEntity.created(location).body(newMusicGenre);
    }

    @PutMapping
    public ResponseEntity<MusicGenre> updateMusicGenre(@RequestBody MusicGenreUpdateDTO musicGenreData) {
        MusicGenre updatedMusicGenre = musicGenreService.updateMusicGenre(musicGenreData);

        return ResponseEntity.ok(updatedMusicGenre);
    }

    @DeleteMapping("/{musicGenreId}")
    public ResponseEntity<Void> deleteMusicGenre(@PathVariable long musicGenreId) {
        musicGenreService.deleteMusicGenre(musicGenreId);

        return ResponseEntity.noContent().build();
    }
}
