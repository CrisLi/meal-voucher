package org.chris.mealvoucher.service;

import org.chris.mealvoucher.entity.Team;
import org.chris.mealvoucher.entity.User;
import org.chris.mealvoucher.exception.AuthException;
import org.chris.mealvoucher.exception.NotFoundException;
import org.chris.mealvoucher.repo.TeamRepo;
import org.chris.mealvoucher.repo.UserRepo;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@Transactional(readOnly = true)
public class UserServiceImpl implements UserService {

    @Autowired
    private UserRepo userRepo;

    @Autowired
    private TeamRepo teamRepo;

    @Override
    @Transactional
    public User createUser(User user, int teamId) {

        Team team = teamRepo.getOne(teamId);

        user.setTeam(team);

        // TODO encrypt password
        return userRepo.save(user);
    }

    @Override
    public User login(String login, String password) {
        return userRepo.findByLoginAndPassword(login, password)
            .orElseThrow(() -> new AuthException());
    }

    @Override
    public User findById(int id) {
        return userRepo.findById(id)
            .orElseThrow(() -> new NotFoundException());
    }

}
