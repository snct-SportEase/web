module.exports = {
    // Conventional Commitsの規約をベースにする
    extends: ['@commitlint/config-conventional'],

    // 独自のルール
    rules: {
        // typeとして許可するリストを設定
        // Level: 2
        'type-enum': [
            2,
            'always',
            [
                'feat',
                'fix',
                'docs',
                'style',
                'ref',
                'perf',
                'test',
                'chore',
                'revert',
                'build',
                'release',
                'ci',
                'init'
            ]
        ],
        'subject-case': [
            0,
            'always',
            'lower-case',
        ]
    }
};
