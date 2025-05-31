package com.undr.demo.dto;

import com.undr.demo.domain.enums.ProjectStatusEnum;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;

import java.time.LocalDate;

@Getter
@Setter
@AllArgsConstructor
public class ProjectCreationDTO {
    String projectName;
    LocalDate projectFoundation;
    ProjectStatusEnum status;
}
