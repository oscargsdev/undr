package com.undr.demo.domain;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
@AllArgsConstructor
@Entity
public class ProjectLocation {

    @Id
    @Column(name = "project_id")
    private Long projectId;

    @OneToOne
    @MapsId
    @JsonIgnore
    private Project project;

    private String hometown;
    private String base;
    private String movingArea;

    public ProjectLocation() {
    }

    public ProjectLocation(String hometown, String base, String movingArea) {
        this.hometown = hometown;
        this.base = base;
        this.movingArea = movingArea;
    }

    @Override
    public String toString() {
        return "ProjectLocation{" +
                "projectId=" + projectId +
                ", hometown='" + hometown + '\'' +
                ", base='" + base + '\'' +
                ", movingArea='" + movingArea + '\'' +
                '}';
    }
}
