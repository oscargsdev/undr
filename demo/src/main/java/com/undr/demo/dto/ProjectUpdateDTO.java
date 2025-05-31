package com.undr.demo.dto;

import com.undr.demo.domain.enums.ProjectStatusEnum;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;

import java.time.LocalDate;

@Getter
@Setter
@AllArgsConstructor
public class ProjectUpdateDTO{
    private Long projectId;
    private String projectName;
    private LocalDate projectFoundation;
    private ProjectStatusEnum status;
}
