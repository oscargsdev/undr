package com.undr.demo.repository;

import com.undr.demo.domain.ProjectLocation;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface ProjectLocationRepository extends JpaRepository<ProjectLocation, Long> {
}
