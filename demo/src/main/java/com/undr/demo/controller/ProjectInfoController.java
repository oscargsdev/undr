package com.undr.demo.controller;

import com.undr.demo.domain.*;
import com.undr.demo.dto.*;
import com.undr.demo.repository.*;
import lombok.RequiredArgsConstructor;
import org.modelmapper.ModelMapper;
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

    @Autowired
    private final MusicGenreRepository musicGenreRepository;

    private final ModelMapper mapper = new ModelMapper();

    @GetMapping("/full-info")
    public ResponseEntity<Iterable<ProjectFullInfoDTO>> getFullInfo(){
        List<ProjectFullInfoDTO> fullInfo = new ArrayList<>();

        List<Project> projects = this.projectRepository.findAll().stream().toList();

        for (Project project : projects){
            long id = project.getProjectId();

            ProjectGenre pg = this.projectGenreRepository.findById(id).orElse(new ProjectGenre());
            MusicGenreDTO mainGenre = mapper.map(musicGenreRepository.findById(
                    pg.getMainGenreId() != null ? pg.getMainGenreId() : 10L
            ).get(), MusicGenreDTO.class);
            List<MusicGenreDTO> subGenres = new ArrayList<>();

            for(Long subGenre : pg.getSubGenresIds()){

                        subGenres.add(mapper.map(this.musicGenreRepository.findById(subGenre != null ? subGenre : 1L).get(), MusicGenreDTO.class));
            }




            ProjectGenreDTO genreDTO = new ProjectGenreDTO();
            genreDTO.setMainGenre(mainGenre);
            genreDTO.setSubGenres(subGenres);
            ProjectLocationDTO locationDTO = mapper.map(this.projectLocationRepository.findById(id).orElse(new ProjectLocation()), ProjectLocationDTO.class);
            StreamingLinksDTO streamingLinksDTO = mapper.map(this.streamingLinksRepository.findById(id).orElse(new StreamingLinks()), StreamingLinksDTO.class);
            SocialLinksDTO socialLinksDTO = mapper.map(this.socialLinksRepository.findById(id).orElse(new SocialLinks()), SocialLinksDTO.class);

            fullInfo.add(new ProjectFullInfoDTO(
                    project,
                    genreDTO,
                    locationDTO,
                    streamingLinksDTO,
                    socialLinksDTO
            ));
        }

        return ResponseEntity.ok(fullInfo.stream().toList());
    }
}
