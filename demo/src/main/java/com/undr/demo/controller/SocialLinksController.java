package com.undr.demo.controller;

import com.undr.demo.domain.SocialLinks;
import com.undr.demo.repository.SocialLinksRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class SocialLinksController {
    @Autowired
    private final SocialLinksRepository socialLinksRepository;

    public SocialLinksController(SocialLinksRepository socialLinksRepository) {
        this.socialLinksRepository = socialLinksRepository;
    }

    @GetMapping("/social-links")
    public ResponseEntity<Iterable<SocialLinks>> getSocialLinks() {
        return ResponseEntity.ok(socialLinksRepository.findAll().stream().toList());
    }
}
