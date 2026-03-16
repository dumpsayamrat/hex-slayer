package com.hexslayer.config;

import com.hexslayer.middleware.SessionAuthInterceptor;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class WebConfig implements WebMvcConfigurer {

    private final SessionAuthInterceptor sessionAuth;

    public WebConfig(SessionAuthInterceptor sessionAuth) {
        this.sessionAuth = sessionAuth;
    }

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(sessionAuth)
                .addPathPatterns("/api/**")
                .excludePathPatterns("/api/health", "/api/player/init");
    }
}
