package com.hexslayer.middleware;

import com.hexslayer.model.Player;
import com.hexslayer.repository.PlayerRepository;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

import java.util.Optional;

@Component
public class SessionAuthInterceptor implements HandlerInterceptor {

    private final PlayerRepository playerRepo;

    public SessionAuthInterceptor(PlayerRepository playerRepo) {
        this.playerRepo = playerRepo;
    }

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
        String header = request.getHeader("Authorization");
        if (header == null || header.isEmpty()) {
            response.setStatus(HttpStatus.UNAUTHORIZED.value());
            response.setContentType("application/json");
            response.getWriter().write("{\"error\":\"missing authorization header\"}");
            return false;
        }

        if (!header.startsWith("Bearer ")) {
            response.setStatus(HttpStatus.UNAUTHORIZED.value());
            response.setContentType("application/json");
            response.getWriter().write("{\"error\":\"invalid authorization format\"}");
            return false;
        }

        String token = header.substring(7);
        Optional<Player> player = playerRepo.findBySessionToken(token);
        if (player.isEmpty()) {
            response.setStatus(HttpStatus.UNAUTHORIZED.value());
            response.setContentType("application/json");
            response.getWriter().write("{\"error\":\"invalid session token\"}");
            return false;
        }

        request.setAttribute("player", player.get());
        return true;
    }
}
