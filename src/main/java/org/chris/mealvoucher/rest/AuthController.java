package org.chris.mealvoucher.rest;

import java.util.Collections;
import java.util.Map;

import javax.servlet.http.HttpSession;

import org.chris.mealvoucher.dto.AuthRequest;
import org.chris.mealvoucher.entity.User;
import org.chris.mealvoucher.exception.AuthException;
import org.chris.mealvoucher.service.UserService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/auth")
public class AuthController {

    @Autowired
    private UserService userService;

    @PostMapping
    public Map<String, Object> auth(@RequestBody AuthRequest auth, HttpSession session) {

        User user = userService.login(auth.getLogin(), auth.getPassword());

        session.setAttribute("currentUserId", user.getId());

        return Collections.singletonMap("token", session.getId());
    }

    @DeleteMapping
    public void logout(HttpSession session) {
        session.invalidate();
    }

    @GetMapping
    public User getAuth(HttpSession session) {

        try {
            int userId = (int) session.getAttribute("currentUserId");
            return userService.findById(userId);
        } catch (Exception e) {
            throw new AuthException();
        }

    }
}
