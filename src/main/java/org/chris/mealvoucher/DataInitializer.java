package org.chris.mealvoucher;

import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.Random;
import java.util.stream.IntStream;

import org.chris.mealvoucher.entity.Team;
import org.chris.mealvoucher.entity.User;
import org.chris.mealvoucher.entity.User.Role;
import org.chris.mealvoucher.entity.Voucher;
import org.chris.mealvoucher.repo.TeamRepo;
import org.chris.mealvoucher.repo.UserRepo;
import org.chris.mealvoucher.repo.VoucherRepo;
import org.chris.mealvoucher.service.UserService;
import org.chris.mealvoucher.service.VoucherService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

@Component
public class DataInitializer implements CommandLineRunner {

    @Autowired
    private VoucherService voucherService;

    @Autowired
    private UserService userService;

    @Autowired
    private UserRepo userRepo;

    @Autowired
    private VoucherRepo voucherRepo;

    @Autowired
    private TeamRepo teamRepo;

    @Override
    public void run(String... args) throws Exception {

        cleanUp();

        initAdmin();

        List<Team> teams = initTeam();

        initVoucher(teams);

        initUsers(teams);

    }

    private void cleanUp() {
        voucherRepo.deleteAll();
        userRepo.deleteAll();
        teamRepo.deleteAll();
    }

    private void initAdmin() {

        Team adminTeam = new Team();

        adminTeam.setName("admin team");

        adminTeam = teamRepo.save(adminTeam);

        User admin = new User();

        admin.setRole(Role.ADMIN);
        admin.setLogin("admin");
        admin.setPassword("password");
        admin.setDisplayName("admin");
        admin.setTeam(adminTeam);

        userService.createUser(admin, adminTeam.getId());
    }

    private List<Team> initTeam() {

        Team team1 = new Team();

        team1.setName("team1");

        team1 = teamRepo.save(team1);

        Team team2 = new Team();

        team2.setName("team2");

        team2 = teamRepo.save(team2);

        return Arrays.asList(team1, team2);
    }

    private void initVoucher(List<Team> teams) {

        Random random = new Random();
        Iterator<Integer> pool = random.ints(15, 30).iterator();

        IntStream.range(8, 13).forEach(i -> {
            teams.forEach(t -> {
                Voucher voucher = new Voucher();
                voucher.setYear(2018);
                voucher.setMonth(i);
                voucher.setQuota(pool.next());
                voucherService.createVoucher(voucher, t.getId());
            });
        });

    }

    private void initUsers(List<Team> teams) {

        User user1 = new User();

        user1.setRole(Role.MEMBER);
        user1.setLogin("chris");
        user1.setPassword("password");
        user1.setDisplayName("Chris Li");
        user1.setTeam(teams.get(0));

        User user2 = new User();

        user2.setRole(Role.MEMBER);
        user2.setLogin("kitty");
        user2.setPassword("password");
        user2.setDisplayName("Kitty Yu");
        user2.setTeam(teams.get(0));

        User leader1 = new User();

        leader1.setRole(Role.LEADER);
        leader1.setLogin("leader1");
        leader1.setPassword("password");
        leader1.setDisplayName("Team 1 Leader");
        leader1.setTeam(teams.get(0));

        Arrays.asList(user1, user2, leader1)
            .forEach(u -> userService.createUser(u, teams.get(0).getId()));

        User user3 = new User();

        user3.setRole(Role.MEMBER);
        user3.setLogin("john");
        user3.setPassword("password");
        user3.setDisplayName("John Doe");
        user3.setTeam(teams.get(1));

        User user4 = new User();

        user4.setRole(Role.MEMBER);
        user4.setLogin("david");
        user4.setPassword("password");
        user4.setDisplayName("David Beck");
        user4.setTeam(teams.get(1));

        User leader2 = new User();

        leader2.setRole(Role.LEADER);
        leader2.setLogin("leader2");
        leader2.setPassword("password");
        leader2.setDisplayName("Team 2 Leader");
        leader2.setTeam(teams.get(1));

        Arrays.asList(user3, user4, leader2)
            .forEach(u -> userService.createUser(u, teams.get(1).getId()));
    }

}
