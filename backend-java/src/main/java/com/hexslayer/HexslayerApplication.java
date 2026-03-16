package com.hexslayer;

import com.hexslayer.game.GameEngine;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;

@SpringBootApplication
public class HexslayerApplication {

	public static void main(String[] args) {
		SpringApplication.run(HexslayerApplication.class, args);
	}

	@Bean
	CommandLineRunner startEngine(GameEngine engine) {
		return args -> engine.start();
	}
}
