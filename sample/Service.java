package com.example.demo;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestTemplate;

@RestController
public class DemoController {

    private final RestTemplate restTemplate = new RestTemplate();

    // Fetches external data
    // Returns a string response
    @GetMapping("/hello")
    public String sayHello() {
        restTemplate.getForObject("http://external-service/api", String.class);
        return "Hello World";
    }
}
