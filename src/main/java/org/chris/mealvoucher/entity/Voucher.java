package org.chris.mealvoucher.entity;

import javax.persistence.Entity;
import javax.persistence.JoinColumn;
import javax.persistence.ManyToOne;
import javax.persistence.Table;

import org.springframework.data.jpa.domain.AbstractPersistable;

import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
@Entity
@Table(name = "vouchers")
public class Voucher extends AbstractPersistable<Integer> {

    @ManyToOne
    @JoinColumn(name = "team_id")
    // @JsonIgnore
    private Team team;

    private int year;
    private int month;
    private int quota;

}
