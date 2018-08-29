package org.chris.mealvoucher.service;

import java.util.List;

import org.chris.mealvoucher.entity.Voucher;

public interface VoucherService {

    public Voucher createVoucher(Voucher voucher, int teamId);

    public Voucher findByTeamIdAndYearAndMonth(int teamId, int year, int month);

    public List<Voucher> findByYearAndMonth(int year, int month);

    public void updateQuota(int id, int quota);

}
