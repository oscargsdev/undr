package com.undr.demo.controller;

import com.undr.demo.domain.StreamingLinks;
import com.undr.demo.repository.StreamingLinksRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class StreamingLinksController {
    @Autowired
    private final StreamingLinksRepository streamingLinksRepository;

    public StreamingLinksController(StreamingLinksRepository streamingLinksRepository){
        this.streamingLinksRepository = streamingLinksRepository;
    }

    @GetMapping("/streaming-links")
    public ResponseEntity<Iterable<StreamingLinks>> getStreamingLinks(){
        return ResponseEntity.ok(streamingLinksRepository.findAll().stream().toList());
    }
}
