package com.undr.demo.domain;

import com.undr.demo.domain.enums.ProjectStatusEnum;
import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.RequiredArgsConstructor;
import org.hibernate.annotations.Cascade;
import org.hibernate.annotations.JdbcTypeCode;
import org.hibernate.type.SqlTypes;

import java.time.LocalDate;
import java.util.Date;
import java.util.Objects;

@Entity
@Data
@NoArgsConstructor(force = true)
@RequiredArgsConstructor
public class Project {

    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private final Long projectId;

    private String projectName;
    private LocalDate projectFoundation;

    @Enumerated(EnumType.STRING)
    private ProjectStatusEnum status;

    @Override
    public boolean equals(Object o) {
        if (o == null || getClass() != o.getClass()) return false;

        Project project = (Project) o;
        return projectId.equals(project.projectId);
    }

    @Override
    public int hashCode() {
        return projectId.hashCode();
    }

    @Override
    public String toString() {
        return "Project{" +
                "projectId=" + projectId +
                ", projectName='" + projectName + '\'' +
                ", projectFoundation=" + projectFoundation +
                ", status=" + status +
//                ", genre=" + genre.toString() +
                '}';
    }
}
