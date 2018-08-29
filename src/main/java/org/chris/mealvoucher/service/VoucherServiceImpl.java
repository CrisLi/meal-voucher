package org.chris.mealvoucher.service;

import java.time.LocalDate;
import java.util.List;

import org.chris.mealvoucher.entity.Team;
import org.chris.mealvoucher.entity.Voucher;
import org.chris.mealvoucher.exception.BadRequestException;
import org.chris.mealvoucher.exception.NotFoundException;
import org.chris.mealvoucher.repo.TeamRepo;
import org.chris.mealvoucher.repo.VoucherRepo;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@Transactional
public class VoucherServiceImpl implements VoucherService {

    @Autowired
    private VoucherRepo voucherRepo;

    @Autowired
    private TeamRepo teamRepo;

    @Value("${meal-voucher.deadline.quota:5}")
    private int deadlineForEditQuota;

    @Override
    public Voucher createVoucher(Voucher voucher, int teamId) {

        Team team = teamRepo.getOne(teamId);

        voucher.setTeam(team);

        return voucherRepo.save(voucher);
    }

    @Override
    @Transactional(readOnly = true)
    public Voucher findByTeamIdAndYearAndMonth(int teamId, int year, int month) {
        return voucherRepo.findByTeamIdAndYearAndMonth(teamId, year, month)
            .orElseThrow(() -> new NotFoundException());
    }

    @Override
    @Transactional(readOnly = true)
    public List<Voucher> findByYearAndMonth(int year, int month) {
        return voucherRepo.findByYearAndMonth(year, month);
    }

    @Override
    public void updateQuota(int id, int quota) {

        LocalDate today = LocalDate.now();

        if (today.getDayOfMonth() > deadlineForEditQuota) {
            throw new BadRequestException();
        }

        voucherRepo.updateQuota(id, quota);
    }

}
