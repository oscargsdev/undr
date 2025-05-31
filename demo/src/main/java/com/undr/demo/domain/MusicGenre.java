package com.undr.demo.domain;

import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Data
@NoArgsConstructor(force = true)
public class MusicGenre {
    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private final Long musicGenreId;

    // TODO: by now only the name is _necessary_. Will think about how to enrich this class.
    String musicGenreName;
}
