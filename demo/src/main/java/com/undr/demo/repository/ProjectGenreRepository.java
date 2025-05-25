package com.undr.demo.repository;

import com.undr.demo.domain.ProjectGenre;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface ProjectGenreRepository extends JpaRepository<ProjectGenre, Long> {
}
