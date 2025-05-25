package com.undr.demo.dto;

import com.undr.demo.domain.*;

public record ProjectFullInfoDTO(Project project, ProjectGenre projectGenre, ProjectLocation projectLocation, StreamingLinks streamingLinks, SocialLinks socialLinks){

}
