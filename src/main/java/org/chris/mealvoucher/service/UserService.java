package org.chris.mealvoucher.service;

import org.chris.mealvoucher.entity.User;

public interface UserService {

    public User createUser(User user, int teamId);

    public User login(String login, String password);
    
    public User findById(int id);

}
