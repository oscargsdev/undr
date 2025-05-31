package com.undr.demo.controller;

import com.undr.demo.domain.ProjectGenre;
import com.undr.demo.repository.ProjectGenreRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class ProjectGenreController {
    @Autowired
    private final ProjectGenreRepository projectGenreRepository;

    public ProjectGenreController(ProjectGenreRepository projectGenreRepository) {
        this.projectGenreRepository = projectGenreRepository;
    }

    @GetMapping("/genres")
    public ResponseEntity<Iterable<ProjectGenre>> getGenres() {
        return ResponseEntity.ok(projectGenreRepository.findAll().stream().toList());
    }
}
