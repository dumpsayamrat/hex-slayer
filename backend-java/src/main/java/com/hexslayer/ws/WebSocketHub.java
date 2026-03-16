package com.hexslayer.ws;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;

import java.io.IOException;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArraySet;
import java.util.function.Consumer;

@Component
public class WebSocketHub {

    private static final Logger log = LoggerFactory.getLogger(WebSocketHub.class);
    private static final ObjectMapper mapper = new ObjectMapper();

    private final Map<String, Set<WebSocketSession>> topics = new ConcurrentHashMap<>();

    private Consumer<String> onSubscribe;

    public void setOnSubscribe(Consumer<String> onSubscribe) {
        this.onSubscribe = onSubscribe;
    }

    public void subscribe(String topic, WebSocketSession session) {
        topics.computeIfAbsent(topic, k -> new CopyOnWriteArraySet<>()).add(session);
        int count = topics.get(topic).size();
        log.info("ws hub: conn subscribed to {} ({} subscribers)", topic, count);

        if (onSubscribe != null) {
            onSubscribe.accept(topic);
        }
    }

    public void unsubscribe(String topic, WebSocketSession session) {
        Set<WebSocketSession> sessions = topics.get(topic);
        if (sessions != null) {
            sessions.remove(session);
            if (sessions.isEmpty()) {
                topics.remove(topic);
            }
        }
    }

    public void unsubscribeAll(WebSocketSession session) {
        for (var entry : topics.entrySet()) {
            entry.getValue().remove(session);
            if (entry.getValue().isEmpty()) {
                topics.remove(entry.getKey());
            }
        }
    }

    public void broadcast(String topic, Object payload) {
        Set<WebSocketSession> sessions = topics.get(topic);
        if (sessions == null || sessions.isEmpty()) return;

        try {
            String json = mapper.writeValueAsString(payload);
            TextMessage msg = new TextMessage(json);
            for (WebSocketSession session : sessions) {
                try {
                    if (session.isOpen()) {
                        synchronized (session) {
                            session.sendMessage(msg);
                        }
                    }
                } catch (IOException e) {
                    log.warn("ws hub: broadcast error on topic {}: {}", topic, e.getMessage());
                }
            }
        } catch (IOException e) {
            log.error("ws hub: failed to serialize payload", e);
        }
    }

    public int subscriberCount(String topic) {
        Set<WebSocketSession> sessions = topics.get(topic);
        return sessions == null ? 0 : sessions.size();
    }
}
