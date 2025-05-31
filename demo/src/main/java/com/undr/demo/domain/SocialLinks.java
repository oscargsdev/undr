package com.undr.demo.domain;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
@Entity
public class SocialLinks {
    @Id
    @Column(name = "project_id")
    private Long projectId;

    @OneToOne
    @MapsId
    @JoinColumn(name = "project_id")
    @JsonIgnore
    private Project project;

    private String instagram;
    private String tiktok;
    private String youtube;
    private String facebook;

    public SocialLinks() {
    }

    public SocialLinks(String instagram, String tiktok, String youtube, String facebook) {
        this.instagram = instagram;
        this.tiktok = tiktok;
        this.youtube = youtube;
        this.facebook = facebook;
    }

    @Override
    public String toString() {
        return "SocialLinks{" +
                "instagram='" + instagram + '\'' +
                ", tiktok='" + tiktok + '\'' +
                ", youtube='" + youtube + '\'' +
                ", facebook='" + facebook + '\'' +
                '}';
    }
}
