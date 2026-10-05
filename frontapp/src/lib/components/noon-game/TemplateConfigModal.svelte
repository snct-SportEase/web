<script>
 let {
  selectedTemplateType, templateConfigForm = $bindable(), classes,
  closeTemplateConfig, saveDefaultGroups, createTemplate,
  isInteractive, creatingTemplate, hasSessionForTemplate, templateKeyFor
 } = $props();
</script>

      <div class="app-layer-modal fixed top-0 left-0 right-0 bottom-0 flex items-center justify-center min-h-screen overflow-y-auto">
        <button
          type="button"
          class="absolute inset-0 bg-black bg-opacity-50"
          onclick={closeTemplateConfig}
          aria-label="テンプレート設定モーダルを閉じる"></button>
        <div
          class="relative bg-white rounded-lg p-6 max-w-2xl w-full mx-4 my-4 max-h-[90vh] overflow-y-auto"
          role="document">
          <div class="flex justify-between items-center mb-4">
            <h2 class="text-2xl font-semibold text-gray-800">
              {#if selectedTemplateType === 'year-relay'}学年対抗リレー
              {:else if selectedTemplateType === 'course-relay'}コース対抗リレー
              {:else if selectedTemplateType === 'tug-of-war'}綱引き
              {:else if selectedTemplateType === 'typing'}競技タイピング
						{:else if selectedTemplateType === 'borrowing-race'}借り物競争
              {/if} テンプレート設定
            </h2>
            <button class="text-gray-500 hover:text-gray-700" onclick={closeTemplateConfig}>×</button>
          </div>

          <div class="space-y-4">
            <div class="border rounded-lg p-4 space-y-4">
              <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">セッション設定</h3>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <label class="flex flex-col text-sm font-medium text-gray-700 md:col-span-2">
                  セッション名
                  <input class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.name} placeholder="例: 学年対抗リレー" />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700 md:col-span-2">
                  説明
                  <textarea class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.description} rows="3" placeholder="概要やメモを入力"></textarea>
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  日時
                  <input type="datetime-local" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.scheduled_at} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  会場
                  <input class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.location} placeholder="例: 視聴覚室" />
                </label>
				{#if selectedTemplateType === 'borrowing-race'}
					<label class="flex flex-col text-sm font-medium text-gray-700">
						競技時間（分）
						<input type="number" min="1" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.duration_minutes} />
					</label>
				{/if}
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  公開状態
                  <select class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.status}>
                    <option value="draft">下書き（学生には非表示）</option>
                    <option value="published">公開（学生に表示）</option>
                  </select>
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  モード
                  <select class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.mode}>
                    <option value="mixed">クラス＆グループ混在</option>
                    <option value="class">クラス対抗のみ</option>
                    <option value="group">グループ対抗のみ</option>
                  </select>
                </label>
                <label class="flex items-center space-x-2 text-sm font-medium text-gray-700">
                  <input type="checkbox" bind:checked={templateConfigForm.allow_manual_points} />
                  <span>手動加点を許可</span>
                </label>
                <label class="flex items-center space-x-2 text-sm font-medium text-gray-700 md:col-span-2">
                  <input type="checkbox" bind:checked={templateConfigForm.exclude_registration_limit} />
                  <span>この昼競技を重複登録の競技数制限から除外する</span>
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  勝利ポイント
                  <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.win_points} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  敗北ポイント
                  <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.loss_points} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  引き分けポイント
                  <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.draw_points} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  参加ポイント
                  <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.participation_points} />
                </label>
              </div>
              {#if selectedTemplateType === 'year-relay'}
                <div class="border rounded-lg p-4 space-y-4 bg-blue-50">
                  <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">点数設定</h3>
                  <p class="text-sm text-gray-600">Aブロック、Bブロック、総合順位の点数を設定します。</p>

                  <div class="space-y-4">
                    <div>
                      <h4 class="text-md font-semibold text-gray-700 mb-2">Aブロック</h4>
                      <div class="grid grid-cols-6 gap-2">
                        {#each [1, 2, 3, 4, 5, 6] as rank (rank)}
                          <label class="flex flex-col text-xs font-medium text-gray-700">
                            {rank}位
                            <input type="number" class="mt-1 border rounded px-2 py-1 text-sm" value={templateConfigForm.year_relay_points.block_a[rank]} oninput={(e) => {
                              templateConfigForm.year_relay_points.block_a = {...templateConfigForm.year_relay_points.block_a, [rank]: Number(e.target.value) || 0};
                            }} />
                          </label>
                        {/each}
                      </div>
                    </div>

                    <div>
                      <h4 class="text-md font-semibold text-gray-700 mb-2">Bブロック</h4>
                      <div class="grid grid-cols-6 gap-2">
                        {#each [1, 2, 3, 4, 5, 6] as rank (rank)}
                          <label class="flex flex-col text-xs font-medium text-gray-700">
                            {rank}位
                            <input type="number" class="mt-1 border rounded px-2 py-1 text-sm" value={templateConfigForm.year_relay_points.block_b[rank]} oninput={(e) => {
                              templateConfigForm.year_relay_points.block_b = {...templateConfigForm.year_relay_points.block_b, [rank]: Number(e.target.value) || 0};
                            }} />
                          </label>
                        {/each}
                      </div>
                    </div>

                    <div>
                      <h4 class="text-md font-semibold text-gray-700 mb-2">総合順位</h4>
                      <div class="grid grid-cols-6 gap-2">
                        {#each [1, 2, 3, 4, 5, 6] as rank (rank)}
                          <label class="flex flex-col text-xs font-medium text-gray-700">
                            {rank}位
                            <input type="number" class="mt-1 border rounded px-2 py-1 text-sm" value={templateConfigForm.year_relay_points.overall[rank]} oninput={(e) => {
                              templateConfigForm.year_relay_points.overall = {...templateConfigForm.year_relay_points.overall, [rank]: Number(e.target.value) || 0};
                            }} />
                          </label>
                        {/each}
                      </div>
                    </div>
                  </div>
                </div>
				{:else if selectedTemplateType === 'borrowing-race'}
					<div class="border rounded-lg p-4 space-y-4 bg-blue-50">
						<h3 class="text-lg font-semibold text-gray-800 border-b pb-2">順位点設定</h3>
						<p class="text-sm text-gray-600">同点は同順位となり、同じ順位点を付与します。未設定の順位は0点です。</p>
						<div class="grid grid-cols-2 md:grid-cols-5 gap-3">
							{#each Array.from({ length: Math.max(10, templateConfigForm.participant_class_ids.length) }, (_, index) => index + 1) as rank (rank)}
								<label class="flex flex-col text-xs font-medium text-gray-700">
									{rank}位
									<input type="number" min="0" class="mt-1 border rounded px-2 py-1 text-sm" value={templateConfigForm.points_by_rank[rank] ?? 0} oninput={(event) => {
										templateConfigForm.points_by_rank = { ...templateConfigForm.points_by_rank, [rank]: Number(event.target.value) || 0 };
									}} />
								</label>
							{/each}
						</div>
					</div>
					<div class="border rounded-lg p-4 space-y-3 bg-green-50">
						<div class="flex items-center justify-between">
							<h3 class="text-lg font-semibold text-gray-800">参加クラス</h3>
							<button type="button" class="text-sm text-indigo-700" onclick={() => { templateConfigForm.participant_class_ids = classes.map((item) => item.id); }}>全クラスを選択</button>
						</div>
						<div class="grid grid-cols-2 md:grid-cols-4 gap-2">
							{#each classes as classItem (classItem.id)}
								<label class="flex items-center gap-2 rounded border bg-white px-3 py-2 text-sm">
									<input type="checkbox" value={classItem.id} checked={templateConfigForm.participant_class_ids.includes(classItem.id)} onchange={(event) => {
										templateConfigForm.participant_class_ids = event.target.checked
											? [...templateConfigForm.participant_class_ids, classItem.id]
											: templateConfigForm.participant_class_ids.filter((id) => id !== classItem.id);
									}} />
									{classItem.name}
								</label>
							{/each}
						</div>
					</div>
              {:else if selectedTemplateType === 'course-relay' || selectedTemplateType === 'tug-of-war' || selectedTemplateType === 'typing'}
                <div class="border rounded-lg p-4 space-y-4 bg-blue-50">
                  <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">点数設定</h3>
                  <p class="text-sm text-gray-600">順位ごとの点数を設定します。</p>
                  <div class="grid grid-cols-3 gap-4">
                    <label class="flex flex-col text-sm font-medium text-gray-700">
                      1位の点数
                      <input type="number" class="mt-1 border rounded px-3 py-2" value={templateConfigForm.points_by_rank[1]} oninput={(e) => {
                        templateConfigForm.points_by_rank = {...templateConfigForm.points_by_rank, 1: Number(e.target.value) || 0};
                      }} />
                    </label>
                    <label class="flex flex-col text-sm font-medium text-gray-700">
                      2位の点数
                      <input type="number" class="mt-1 border rounded px-3 py-2" value={templateConfigForm.points_by_rank[2]} oninput={(e) => {
                        templateConfigForm.points_by_rank = {...templateConfigForm.points_by_rank, 2: Number(e.target.value) || 0};
                      }} />
                    </label>
                    <label class="flex flex-col text-sm font-medium text-gray-700">
                      3位の点数
                      <input type="number" class="mt-1 border rounded px-3 py-2" value={templateConfigForm.points_by_rank[3]} oninput={(e) => {
                        templateConfigForm.points_by_rank = {...templateConfigForm.points_by_rank, 3: Number(e.target.value) || 0};
                      }} />
                    </label>
                    {#if selectedTemplateType !== 'typing'}
                    <label class="flex flex-col text-sm font-medium text-gray-700">
                      4位の点数
                      <input type="number" class="mt-1 border rounded px-3 py-2" value={templateConfigForm.points_by_rank[4]} oninput={(e) => {
                        templateConfigForm.points_by_rank = {...templateConfigForm.points_by_rank, 4: Number(e.target.value) || 0};
                      }} />
                    </label>
                    {/if}
                  </div>
                </div>
              {/if}

			{#if selectedTemplateType !== 'borrowing-race'}
              <!-- グループ設定 -->
              <div class="border rounded-lg p-4 space-y-4 bg-green-50">
                <div class="flex justify-between items-center">
                  <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">グループ設定</h3>
                  <button
                    class="text-sm px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600"
                    onclick={() => {
                      const templateKeyMap = {
                        'year-relay': 'year_relay',
                        'course-relay': 'course_relay',
                        'tug-of-war': 'tug_of_war',
                        'typing': 'typing'
                      };
                      saveDefaultGroups(templateKeyMap[selectedTemplateType]);
                    }}>
                    デフォルト設定を保存
                  </button>
                </div>
                <p class="text-sm text-gray-600">各グループの名前と所属クラスを設定します。デフォルト設定として保存することもできます。</p>

                <div class="space-y-3">
                  {#each templateConfigForm.groups as group, index (index)}
                    <div class="border rounded p-3 bg-white">
                      <div class="flex items-center space-x-2 mb-2">
                        <label class="flex-1 flex flex-col text-sm font-medium text-gray-700">
                          グループ名
                          <input
                            type="text"
                            class="mt-1 border rounded px-2 py-1 text-sm"
                            value={group.group_name}
                            oninput={(e) => {
                              templateConfigForm.groups[index].group_name = e.target.value;
                              templateConfigForm.groups = [...templateConfigForm.groups];
                            }}
                            placeholder="例: 1年生" />
                        </label>
                        <button
                          class="text-red-500 hover:text-red-700 px-2 py-1 text-sm"
                          onclick={() => {
                            templateConfigForm.groups = templateConfigForm.groups.filter((_, i) => i !== index);
                          }}>
                          削除
                        </button>
                      </div>
                      <div class="flex flex-col space-y-2">
                        <label class="text-sm font-medium text-gray-700" for={`group-class-names-${index}`}>クラス名（カンマ区切り）</label>
                        <input
                          id={`group-class-names-${index}`}
                          type="text"
                          class="border rounded px-2 py-1 text-sm"
                          value={group.class_names ? group.class_names.join(', ') : ''}
                          oninput={(e) => {
                            const classNames = e.target.value.split(',').map(s => s.trim()).filter(s => s);
                            templateConfigForm.groups[index].class_names = classNames;
                            templateConfigForm.groups = [...templateConfigForm.groups];
                          }}
                          placeholder="例: 1-1, 1-2, 1-3" />
                        {#if group.class_names && group.class_names.length > 0}
                          <div class="flex flex-wrap gap-1">
                            {#each group.class_names as className (className)}
                              <span class="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs">{className}</span>
                            {/each}
                          </div>
                        {/if}
                      </div>
                    </div>
                  {/each}
                  <button
                    class="w-full px-4 py-2 border-2 border-dashed border-gray-300 rounded text-gray-600 hover:border-gray-400 hover:text-gray-800"
                    onclick={() => {
                      templateConfigForm.groups = [...templateConfigForm.groups, { group_name: '', class_names: [] }];
                    }}>
                    + グループを追加
                  </button>
                </div>
              </div>
			{/if}
            </div>

            <div class="flex justify-end space-x-3">
              <button
                class="px-4 py-2 border rounded text-gray-700 hover:bg-gray-50"
                onclick={closeTemplateConfig}>
                キャンセル
              </button>
              <button
                class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                onclick={createTemplate}
                disabled={!isInteractive || creatingTemplate[selectedTemplateType]}>
                {#if creatingTemplate[selectedTemplateType]}
                  {hasSessionForTemplate(templateKeyFor(selectedTemplateType)) ? '更新中...' : '作成中...'}
                {:else}
                  {hasSessionForTemplate(templateKeyFor(selectedTemplateType)) ? '競技を更新' : '昼競技を作成'}
                {/if}
              </button>
            </div>
          </div>
        </div>
      </div>
