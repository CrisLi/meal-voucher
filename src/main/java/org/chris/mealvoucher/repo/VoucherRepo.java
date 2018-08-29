package org.chris.mealvoucher.repo;

import java.util.List;
import java.util.Optional;

import org.chris.mealvoucher.entity.Voucher;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;

public interface VoucherRepo extends JpaRepository<Voucher, Integer> {

    public Optional<Voucher> findByTeamIdAndYearAndMonth(int teamId, int year, int month);

    public List<Voucher> findByYearAndMonth(int year, int month);

    @Modifying
    @Query("update Voucher v set v.quota = :quota where v.id = :id")
    public int updateQuota(int id, int quota);

}
