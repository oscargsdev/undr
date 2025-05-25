package com.undr.demo.controller;

import com.undr.demo.domain.ProjectLocation;
import com.undr.demo.repository.ProjectLocationRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class ProjectLocationController {
    @Autowired
    private final ProjectLocationRepository projectLocationRepository;

    public ProjectLocationController(ProjectLocationRepository projectLocationRepository){
        this.projectLocationRepository = projectLocationRepository;
    }

    @GetMapping("/locations")
    public ResponseEntity<Iterable<ProjectLocation>> getProjectLocations(){
        return ResponseEntity.ok(projectLocationRepository.findAll().stream().toList());
    }
}
