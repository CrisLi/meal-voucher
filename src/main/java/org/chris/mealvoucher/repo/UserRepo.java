package org.chris.mealvoucher.repo;

import java.util.Optional;

import org.chris.mealvoucher.entity.User;
import org.springframework.data.jpa.repository.JpaRepository;

public interface UserRepo extends JpaRepository<User, Integer> {

    public Optional<User> findByLoginAndPassword(String login, String password);

}
