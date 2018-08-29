package org.chris.mealvoucher.rest;

import java.util.List;

import org.chris.mealvoucher.entity.Voucher;
import org.chris.mealvoucher.service.VoucherService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/vouchers")
public class VoucherController {

    @Autowired
    private VoucherService voucherService;

    @GetMapping("/team")
    public Voucher getTeamVoucher(@RequestParam int teamId, @RequestParam int year, @RequestParam int month) {
        return voucherService.findByTeamIdAndYearAndMonth(teamId, year, month);
    }

    @GetMapping
    public List<Voucher> getAll(@RequestParam int year, @RequestParam int month) {
        return voucherService.findByYearAndMonth(year, month);
    }

    @PostMapping("/{id}/change_quota")
    public void changeQuota(@PathVariable("id") int voucherId, @RequestParam("quota") int quota) {
        voucherService.updateQuota(voucherId, quota);
    }

}
