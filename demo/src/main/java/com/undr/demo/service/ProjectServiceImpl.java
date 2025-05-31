package com.undr.demo.service;

import com.undr.demo.domain.Project;
import com.undr.demo.dto.ProjectCreationDTO;
import com.undr.demo.dto.ProjectUpdateDTO;
import com.undr.demo.repository.ProjectRepository;
import org.modelmapper.ModelMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class ProjectServiceImpl implements ProjectService{
    @Autowired
    private final ProjectRepository projectRepository;

    private final ModelMapper mapper = new ModelMapper();

    public ProjectServiceImpl(ProjectRepository projectRepository) {
        this.projectRepository = projectRepository;
    }

    @Override
    public Project createProject(ProjectCreationDTO projectData) {
        Project newProject = mapper.map(projectData, Project.class);
        projectRepository.save(newProject);
        return newProject;
    }

    @Override
    public Project updateProject(ProjectUpdateDTO projectData) {
        Project updatedProject = projectRepository.findById(projectData.getProjectId()).get();

        mapper.map(projectData, updatedProject);
        projectRepository.save(updatedProject);

        return updatedProject;
    }

    @Override
    public Project getProject(Long projectId) {
        Project project = projectRepository.findById(projectId).get();

        return project;
    }

    @Override
    public void deleteProject(Long projectId) {
        Project deletedProject = projectRepository.findById(projectId).get();

        projectRepository.delete(deletedProject);
    }

    @Override
    public Iterable<Project> getAllProjects() {
        return this.projectRepository.findAll().stream().toList();
    }
}
