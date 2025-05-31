package com.undr.demo.service;

import com.undr.demo.domain.Project;
import com.undr.demo.dto.ProjectCreationDTO;
import com.undr.demo.dto.ProjectUpdateDTO;

public interface ProjectService {
    Project createProject(ProjectCreationDTO projectData);
    Project updateProject(ProjectUpdateDTO projectData);
    Project getProject(Long projectId);
    void deleteProject(Long projectId);
    Iterable<Project> getAllProjects();
}
