package com.undr.demo.service;

import com.undr.demo.domain.Project;
import com.undr.demo.dto.ProjectCreationDTO;
import com.undr.demo.dto.ProjectUpdateDTO;

public interface ProjectService {
    Project createProject(ProjectCreationDTO projectData);
    Project updateProject(ProjectUpdateDTO projectData);
    Project getProjectById(long projectId);
    void deleteProject(long projectId);
    Iterable<Project> getAllProjects();
}
