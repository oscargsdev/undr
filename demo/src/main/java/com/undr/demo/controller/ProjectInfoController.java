package com.undr.demo.controller;

import com.undr.demo.domain.*;
import com.undr.demo.dto.ProjectFullInfoDTO;
import com.undr.demo.repository.*;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.ArrayList;
import java.util.List;

@RestController
@RequiredArgsConstructor
public class ProjectInfoController {
    @Autowired
    private final ProjectRepository projectRepository;

    @Autowired
    private final ProjectGenreRepository projectGenreRepository;

    @Autowired
    private final ProjectLocationRepository projectLocationRepository;

    @Autowired
    private final StreamingLinksRepository streamingLinksRepository;

    @Autowired
    private final SocialLinksRepository socialLinksRepository;

    @GetMapping("/full-info")
    public ResponseEntity<Iterable<ProjectFullInfoDTO>> getFullInfo(){
        List<ProjectFullInfoDTO> fullInfo = new ArrayList<>();

        List<Project> projects = this.projectRepository.findAll().stream().toList();

        for (Project project : projects){
            long id = project.getProjectId();

            fullInfo.add(new ProjectFullInfoDTO(
                    project,
                    this.projectGenreRepository.findById(id).orElse(new ProjectGenre()),
                    this.projectLocationRepository.findById(id).orElse(new ProjectLocation()),
                    this.streamingLinksRepository.findById(id).orElse(new StreamingLinks()),
                    this.socialLinksRepository.findById(id).orElse(new SocialLinks())
            ));
        }

        return ResponseEntity.ok(fullInfo.stream().toList());
    }
}
