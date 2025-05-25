package com.undr.demo.controller;

import com.undr.demo.domain.Project;
import com.undr.demo.repository.ProjectRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class ProjectController {
    @Autowired
    private final ProjectRepository projectRepository;

    public ProjectController(ProjectRepository projectRepository){
        this.projectRepository = projectRepository;
    }

    @GetMapping("/projects")
    public ResponseEntity<Iterable<Project>> getProjects(){
        return ResponseEntity.ok(projectRepository.findAll().stream().toList());
    }
}
