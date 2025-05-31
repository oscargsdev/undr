package com.undr.demo.repository;

import com.undr.demo.domain.MusicGenre;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface MusicGenreRepository extends JpaRepository<MusicGenre, Long> {
}
