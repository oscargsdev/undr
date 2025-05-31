package com.undr.demo.domain;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.Getter;
import lombok.Setter;

import java.util.Arrays;

@Entity
@Getter
@Setter
public class ProjectGenre {
    @Id
    private Long projectId;

    @OneToOne(fetch = FetchType.LAZY)
    @MapsId
    @JsonIgnore
    private Project project;

    private Long mainGenreId;
    private Long[] subGenresIds;

    public ProjectGenre() {
    }

    public ProjectGenre(Long mainGenreId, Long... subGenresIds) {
        this.mainGenreId = mainGenreId;
        this.subGenresIds = subGenresIds;
    }

    @Override
    public String toString() {
        return "ProjectGenre{" +
                "projectId=" + projectId +
                ", mainGenre='" + mainGenreId + '\'' +
                ", subGenres=" + Arrays.toString(subGenresIds) +
                '}';
    }
}
