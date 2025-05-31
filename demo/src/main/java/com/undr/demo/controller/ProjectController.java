package com.undr.demo.controller;

import com.undr.demo.domain.Project;
import com.undr.demo.dto.ProjectCreationDTO;
import com.undr.demo.dto.ProjectUpdateDTO;
import com.undr.demo.service.ProjectService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.support.ServletUriComponentsBuilder;

import java.net.URI;

@RestController()
@RequestMapping("/v1/projects")
public class ProjectController {
   @Autowired
   private final ProjectService projectService;

    public ProjectController(ProjectService projectService) {
        this.projectService = projectService;
    }

    @GetMapping()
    public ResponseEntity<Iterable<Project>> getAllProjects(){
        return ResponseEntity.ok(projectService.getAllProjects());
    }

    @GetMapping("/{projectId}")
    public ResponseEntity<Project> getProjectById(@PathVariable long projectId){
        Project project = projectService.getProject(projectId);

        return ResponseEntity.ok(project);
    }

    @PostMapping
    public ResponseEntity<Project> createProject(@RequestBody ProjectCreationDTO projectData){
        Project newProject = projectService.createProject(projectData);
        URI location = ServletUriComponentsBuilder
                .fromCurrentRequest()
                .path("/{userId}")
                .buildAndExpand(newProject.getProjectId())
                .toUri();

        return ResponseEntity.created(location).body(newProject);
    }

    @PutMapping
    public ResponseEntity<Project> updateProject(@RequestBody ProjectUpdateDTO projectData){
        Project updatedProject = projectService.updateProject(projectData);

        return ResponseEntity.ok(updatedProject);
    }

    @DeleteMapping("/{projectId}")
    public ResponseEntity<Void> deleteProject(@PathVariable long projectId){
        projectService.deleteProject(projectId);

        return ResponseEntity.noContent().build();
    }
}
