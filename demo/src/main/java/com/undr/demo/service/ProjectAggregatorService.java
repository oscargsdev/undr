package com.undr.demo.service;

import com.undr.demo.domain.Project;
import com.undr.demo.dto.ProjectFullInfoDTO;

public interface ProjectAggregatorService {
    ProjectFullInfoDTO getProjectFullInfo(Project project);
}
