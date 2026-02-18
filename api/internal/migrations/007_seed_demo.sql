-- +goose Up

-- Demo client
INSERT INTO clients (id, name, email)
VALUES ('c0000000deadbeef', 'PlayPeru Demo', 'demo@playperu.com');

-- Lima Centro Historico scenario with 4 stages
INSERT INTO scenarios (id, name, city, description, stages)
VALUES (
    's0000000deadbeef',
    'Lima Centro Historico',
    'Lima',
    'Explore the historic center of Lima through four iconic landmarks.',
    json('[
        {
            "stageNumber": 1,
            "location": "Plaza Mayor",
            "clue": "Head to the main square where Pizarro founded the city. Look for the bronze fountain in the center.",
            "question": "What year was the fountain in Plaza Mayor built?",
            "correctAnswer": "1651",
            "lat": -12.0464,
            "lng": -77.0300
        },
        {
            "stageNumber": 2,
            "location": "Iglesia de San Francisco",
            "clue": "Walk south to the yellow church with famous underground tunnels.",
            "question": "What are the underground tunnels beneath San Francisco called?",
            "correctAnswer": "catacombs",
            "lat": -12.0463,
            "lng": -77.0275
        },
        {
            "stageNumber": 3,
            "location": "Jiron de la Union",
            "clue": "Stroll down Limas most famous pedestrian street. Find the statue of the liberator.",
            "question": "Which liberator has a statue on Jiron de la Union?",
            "correctAnswer": "San Martin",
            "lat": -12.0500,
            "lng": -77.0350
        },
        {
            "stageNumber": 4,
            "location": "Parque de la Muralla",
            "clue": "Follow the old city wall to the park along the Rimac river.",
            "question": "What century were the original city walls built in?",
            "correctAnswer": "17th",
            "lat": -12.0450,
            "lng": -77.0260
        }
    ]')
);

-- Active game
INSERT INTO games (id, scenario_id, client_id, status, started_at, timer_minutes)
VALUES (
    'g0000000deadbeef',
    's0000000deadbeef',
    'c0000000deadbeef',
    'active',
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
    120
);

-- Two teams
INSERT INTO teams (id, game_id, name, join_token)
VALUES
    ('t000000000incas', 'g0000000deadbeef', 'Los Incas', 'incas-2025'),
    ('t00000000condor', 'g0000000deadbeef', 'Los Condores', 'condores-2025');

-- +goose Down
DELETE FROM teams WHERE game_id = 'g0000000deadbeef';
DELETE FROM games WHERE id = 'g0000000deadbeef';
DELETE FROM scenarios WHERE id = 's0000000deadbeef';
DELETE FROM clients WHERE id = 'c0000000deadbeef';
