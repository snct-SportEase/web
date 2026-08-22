<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import ActionCard from './ActionCard.svelte';
  import Alert from './Alert.svelte';
  import Badge from './Badge.svelte';
  import Button from './Button.svelte';
  import Card from './Card.svelte';
  import DataTable from './DataTable.svelte';
  import EmptyState from './EmptyState.svelte';
  import FormField from './FormField.svelte';
  import LoadingState from './LoadingState.svelte';
  import Modal from './Modal.svelte';
  import ModalFooter from './ModalFooter.svelte';
  import PageHeader from './PageHeader.svelte';

  const { Story } = defineMeta({
    title: 'Components',
    component: Button,
    tags: ['autodocs']
  });
</script>

<Story name="ActionCard" asChild>
  <div class="max-w-sm">
    <ActionCard title="競技要項" description="大会の競技要項を確認できます。" />
  </div>
</Story>

<Story name="Alert" asChild>
  <div class="space-y-3">
    <Alert variant="info">お知らせを確認してください。</Alert>
    <Alert variant="success">保存しました。</Alert>
    <Alert variant="warning">入力内容を確認してください。</Alert>
    <Alert variant="error" title="エラー">保存に失敗しました。</Alert>
  </div>
</Story>

<Story name="Badge" asChild>
  <div class="flex flex-wrap gap-2">
    <Badge>アーカイブ</Badge><Badge variant="info">予定</Badge><Badge variant="success">開催中</Badge><Badge variant="warning">準備中</Badge><Badge variant="danger">無効</Badge>
  </div>
</Story>

<Story name="Button" asChild>
  <div class="flex flex-wrap gap-3">
    <Button>保存</Button><Button variant="secondary">キャンセル</Button><Button variant="danger">削除</Button><Button loading loadingLabel="保存中...">保存</Button><Button disabled>無効</Button>
  </div>
</Story>

<Story name="Card" asChild>
  <div class="grid max-w-xl gap-4 sm:grid-cols-2">
    <Card class="p-5"><h2 class="font-semibold">通常カード</h2><p class="mt-1 text-sm text-gray-600">情報を表示します。</p></Card>
    <Card variant="featured" class="p-5"><h2 class="font-semibold text-indigo-700">注目カード</h2><p class="mt-1 text-sm text-gray-600">重要な情報を表示します。</p></Card>
  </div>
</Story>

<Story name="DataTable" asChild>
  <DataTable>
    {#snippet header()}<tr><th class="px-4 py-3 text-left">クラス</th><th class="px-4 py-3 text-left">出席数</th></tr>{/snippet}
    <tr><td class="px-4 py-3">1年A組</td><td class="px-4 py-3">32名</td></tr>
    <tr><td class="px-4 py-3">1年B組</td><td class="px-4 py-3">30名</td></tr>
  </DataTable>
</Story>

<Story name="EmptyState" asChild>
  <EmptyState title="表示できるデータがありません" description="条件を変更して再度お試しください。" />
</Story>

<Story name="FormField" asChild>
  <div class="max-w-md space-y-4">
    <FormField label="大会名" inputId="event-name" description="参加者に表示される名称です。"><input id="event-name" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" value="春季スポーツ大会" /></FormField>
    <FormField label="シーズン" inputId="season"><select id="season" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"><option>春</option><option>秋</option></select></FormField>
  </div>
</Story>

<Story name="LoadingState" asChild>
  <LoadingState />
</Story>

<Story name="Modal" asChild>
  <Modal open title="大会を作成"><p class="mt-4 text-sm text-gray-600">必要な情報を入力してください。</p>{#snippet footer()}<ModalFooter onconfirm={() => {}} oncancel={() => {}} />{/snippet}</Modal>
</Story>

<Story name="ModalFooter" asChild>
  <div class="flex max-w-lg flex-row-reverse bg-gray-50 p-4"><ModalFooter onconfirm={() => {}} oncancel={() => {}} /></div>
</Story>

<Story name="PageHeader" asChild>
  <PageHeader title="大会情報登録・管理" description="大会の作成と設定変更を行います。"><Button size="sm">新規作成</Button></PageHeader>
</Story>
