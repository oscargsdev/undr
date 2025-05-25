package com.undr.demo.repository;

import com.undr.demo.domain.StreamingLinks;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface StreamingLinksRepository extends JpaRepository<StreamingLinks, Long> {
}
