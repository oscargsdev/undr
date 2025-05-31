package com.undr.demo.domain;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.Builder;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@Getter
@Setter

@Entity
public class StreamingLinks {
    @Id
    @Column(name = "project_id")
    private Long projectId;

    @OneToOne
    @MapsId
    @JoinColumn(name = "project_id")
    @JsonIgnore
    private Project project;

    private String spotify;
    private String tidal;
    private String appleMusic;
    private String bandcamp;
    private String soundcloud;

    public StreamingLinks() {
    }

    public StreamingLinks(String spotify, String tidal, String appleMusic, String bandcamp, String soundcloud) {
        this.spotify = spotify;
        this.tidal = tidal;
        this.appleMusic = appleMusic;
        this.bandcamp = bandcamp;
        this.soundcloud = soundcloud;
    }

    @Override
    public String toString() {
        return "StreamingLinks{" +
                "spotify='" + spotify + '\'' +
                ", tidal='" + tidal + '\'' +
                ", appleMusic='" + appleMusic + '\'' +
                ", bandcamp='" + bandcamp + '\'' +
                ", soundcloud='" + soundcloud + '\'' +
                '}';
    }
}
