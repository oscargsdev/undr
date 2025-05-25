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

    private String mainGenre;
    private String[] subGenres;

    public ProjectGenre(){}
    public ProjectGenre(String mainGenre, String... subGenres){
        this.mainGenre = mainGenre;
        this.subGenres = subGenres;
    }

    @Override
    public String toString() {
        return "ProjectGenre{" +
                "projectId=" + projectId +
                ", mainGenre='" + mainGenre + '\'' +
                ", subGenres=" + Arrays.toString(subGenres) +
                '}';
    }
}
