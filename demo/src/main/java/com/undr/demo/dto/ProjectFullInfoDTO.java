package com.undr.demo.dto;

import com.undr.demo.domain.*;

public record ProjectFullInfoDTO(
        Project project,
        ProjectGenreDTO projectGenre,
        ProjectLocationDTO projectLocation,
        StreamingLinksDTO streamingLinks,
        SocialLinksDTO socialLinks)
{
}
