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

    String musicGenreName;
}
