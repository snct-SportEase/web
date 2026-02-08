<script>
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  // URLパラメータからエラーメッセージをチェック
  onMount(() => {
    if (browser) {
      const urlParams = new URLSearchParams(window.location.search);
      const error = urlParams.get('error');
      
      if (error) {
        // エラーメッセージに基づいて適切なメッセージを表示
        let errorMessage = '';
        switch (error) {
          case 'domain_not_allowed':
            errorMessage = '許可されていないドメインです。@sendai-nct.jpまたは@sendai-nct.ac.jpのメールアドレスを使用してください。';
            break;
          case 'email_not_whitelisted':
            errorMessage = 'このメールアドレスはアクセスが許可されていません。管理者に連絡して、ホワイトリストへの追加を依頼してください。';
            break;
          case 'access_denied':
            errorMessage = 'アクセスが拒否されました。詳しくは管理者にお問い合わせください。';
            break;
          default:
            errorMessage = 'ログインに失敗しました。もう一度お試しください。';
        }
        
        // アラートポップアップを表示
        alert(errorMessage);
        
        // URLからエラーパラメータを削除
        const newUrl = new URL(window.location);
        newUrl.searchParams.delete('error');
        window.history.replaceState({}, '', newUrl);
      }
    }
  });
</script>

<!-- ヒーローセクション -->
<section id="home" class="bg-gradient-to-r from-[#4a69bd] to-[#6a89cc] text-white text-center py-20 md:py-28 px-4 sm:px-8">
    <div class="container mx-auto">
        <h1 class="text-3xl md:text-5xl font-bold mb-4">仙台高専広瀬キャンパス 行事委員会</h1>
        <p class="text-base md:text-xl opacity-95">学校を、もっと面白く。君のアイデアで、最高のイベントを創ろう！</p>
    </div>
</section>

<!-- Aboutセクション -->
<section id="about" class="py-16 md:py-20 bg-white">
    <div class="container mx-auto px-4 sm:px-6 lg:px-8">
        <h2 class="text-3xl md:text-4xl font-bold text-center mb-8 text-gray-900">行事委員会へようこそ！</h2>
        <p class="text-center max-w-4xl mx-auto text-lg md:text-xl leading-relaxed text-gray-700">
          私たち行事委員会は、学校生活を彩る様々なイベントを企画・運営しています。
          仲間と協力し、一つの目標に向かって全力で取り組む。その経験は、きっとあなたの大きな財産になります。
        </p>
    </div>
</section>

<!-- ログインセクション -->
<section id="login" class="py-16 md:py-20 bg-gray-50">
    <div class="container mx-auto px-4 sm:px-6 lg:px-8">
        <div class="max-w-md mx-auto bg-white shadow-lg rounded-xl p-8 md:p-10">
            <div class="flex justify-center mb-6">
                <img src="/icon.png" alt="SportEase Logo" class="h-20 w-20 md:h-24 md:w-24">
            </div>
            <h2 class="text-2xl md:text-3xl font-bold text-center text-gray-900 mb-2">SportEaseへようこそ</h2>
            <p class="text-center text-gray-600 mb-8">学校のGoogleアカウントでサインインしてください。</p>
            <a 
                href="/api/auth/google/login"
                class="w-full flex items-center justify-center bg-white border-2 border-gray-300 text-gray-700 py-3 px-6 rounded-lg hover:bg-gray-50 hover:border-[#4a69bd] hover:text-[#4a69bd] focus:outline-none focus:ring-2 focus:ring-[#4a69bd] focus:ring-offset-2 transition-all duration-200 font-medium shadow-sm"
            >
                <svg class="w-5 h-5 mr-3" aria-hidden="true" focusable="false" data-prefix="fab" data-icon="google" role="img" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 488 512"><path fill="currentColor" d="M488 261.8C488 403.3 381.5 512 244 512 109.8 512 0 402.2 0 261.8S109.8 11.6 244 11.6C318.3 11.6 382.8 45 427.3 99.9L353.5 168.3C327.2 145.3 289.3 129.8 244 129.8c-66.8 0-121.5 54.9-121.5 122.1s54.7 122.1 121.5 122.1c76.3 0 104.5-54.7 108.8-82.9H244v-66.8h236.1c2.4 12.6 3.9 26.1 3.9 40.9z"></path></svg>
                Googleでサインイン
            </a>
        </div>
    </div>
</section>

<!-- イベントセクション -->
<section id="events" class="py-16 md:py-20 bg-white">
    <div class="container mx-auto px-4 sm:px-6 lg:px-8">
        <h2 class="text-3xl md:text-4xl font-bold text-center mb-12 text-gray-900">主なイベント</h2>
        
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-10">
            <div class="bg-white p-8 md:p-10 rounded-xl shadow-lg border border-gray-100 hover:-translate-y-2 hover:shadow-2xl transition-all duration-300">
                <h3 class="text-2xl md:text-3xl font-bold mb-6 text-[#4a69bd]">スポーツ大会</h3>
                <div class="flex flex-wrap gap-2 mb-8">
                    <span class="bg-[#e0e7ff] text-[#4a69bd] px-4 py-2 rounded-full text-sm font-medium">春季スポーツ大会</span>
                    <span class="bg-[#e0e7ff] text-[#4a69bd] px-4 py-2 rounded-full text-sm font-medium">秋季スポーツ大会</span>
                </div>
                <h4 class="text-xl md:text-2xl font-bold mb-6 text-gray-900">運営の流れ</h4>
                <div class="relative border-l-4 border-[#4a69bd] pl-8 ml-4">
                    <div class="relative pb-6 leading-relaxed last:pb-0 before:content-[''] before:absolute before:-left-[2.4rem] before:top-[0.2rem] before:w-6 before:h-6 before:rounded-full before:bg-[#4a69bd] before:border-4 before:border-white before:shadow-md">
                        <p class="text-gray-700">前日に大会準備を進める</p>
                    </div>
                    <div class="relative pb-6 leading-relaxed last:pb-0 before:content-[''] before:absolute before:-left-[2.4rem] before:top-[0.2rem] before:w-6 before:h-6 before:rounded-full before:bg-[#4a69bd] before:border-4 before:border-white before:shadow-md">
                        <p class="text-gray-700">大会当日の朝に最終的な準備をし開会式をする</p>
                    </div>
                    <div class="relative pb-6 leading-relaxed last:pb-0 before:content-[''] before:absolute before:-left-[2.4rem] before:top-[0.2rem] before:w-6 before:h-6 before:rounded-full before:bg-[#4a69bd] before:border-4 before:border-white before:shadow-md">
                        <p class="text-gray-700">大会結果を集計し、閉会式で結果発表を行う</p>
                    </div>
                    <div class="relative pb-6 leading-relaxed last:pb-0 before:content-[''] before:absolute before:-left-[2.4rem] before:top-[0.2rem] before:w-6 before:h-6 before:rounded-full before:bg-[#4a69bd] before:border-4 before:border-white before:shadow-md">
                        <p class="text-gray-700">後日の委員会で反省点を共有し、のちの大会運営に生かす</p>
                    </div>
                </div>
            </div>

            <div class="bg-white p-8 md:p-10 rounded-xl shadow-lg border border-gray-100 hover:-translate-y-2 hover:shadow-2xl transition-all duration-300">
                <h3 class="text-2xl md:text-3xl font-bold mb-6 text-[#4a69bd]">その他のイベント</h3>
                <div class="flex flex-wrap gap-2 mb-8">
                    <span class="bg-[#e0e7ff] text-[#4a69bd] px-4 py-2 rounded-full text-sm font-medium">予餞会</span>
                    <span class="bg-[#e0e7ff] text-[#4a69bd] px-4 py-2 rounded-full text-sm font-medium">新入生歓迎会</span>
                </div>
                <h4 class="text-xl md:text-2xl font-bold mb-4 text-gray-900">予餞会運営の流れ</h4>
                <p class="mb-6 text-gray-700 leading-relaxed">予餞会では、各部活動や対象の教員からビデオレターを募りそれを流すという形式をとっている</p>
                
                <div class="relative border-l-4 border-[#4a69bd] pl-8 ml-4">
                    <div class="relative pb-6 leading-relaxed last:pb-0 before:content-[''] before:absolute before:-left-[2.4rem] before:top-[0.2rem] before:w-6 before:h-6 before:rounded-full before:bg-[#4a69bd] before:border-4 before:border-white before:shadow-md">
                        <p class="text-gray-700">部活動と対象となる教員を整理する</p>
                    </div>
                    <div class="relative pb-6 leading-relaxed last:pb-0 before:content-[''] before:absolute before:-left-[2.4rem] before:top-[0.2rem] before:w-6 before:h-6 before:rounded-full before:bg-[#4a69bd] before:border-4 before:border-white before:shadow-md">
                        <p class="text-gray-700">委員でグループを作り、それぞれいくつかの部活動等を担当する</p>
                    </div>
                    <div class="relative pb-6 leading-relaxed last:pb-0 before:content-[''] before:absolute before:-left-[2.4rem] before:top-[0.2rem] before:w-6 before:h-6 before:rounded-full before:bg-[#4a69bd] before:border-4 before:border-white before:shadow-md">
                        <p class="text-gray-700">担当対象の部活動や教員にビデオレターをお願いする</p>
                    </div>
                    <div class="relative pb-6 leading-relaxed last:pb-0 before:content-[''] before:absolute before:-left-[2.4rem] before:top-[0.2rem] before:w-6 before:h-6 before:rounded-full before:bg-[#4a69bd] before:border-4 before:border-white before:shadow-md">
                        <p class="text-gray-700">ビデオレターを回収し、まとめる</p>
                    </div>
                    <div class="relative pb-6 leading-relaxed last:pb-0 before:content-[''] before:absolute before:-left-[2.4rem] before:top-[0.2rem] before:w-6 before:h-6 before:rounded-full before:bg-[#4a69bd] before:border-4 before:border-white before:shadow-md">
                        <p class="text-gray-700">予餞会当日にまとめてビデオレターを流す</p>
                    </div>
                </div>
            </div>
        </div>
    </div>
</section>

<!-- 役職紹介セクション -->
<section id="roles" class="py-16 md:py-20 bg-gray-50">
    <div class="container mx-auto px-4 sm:px-6 lg:px-8">
        <h2 class="text-3xl md:text-4xl font-bold text-center mb-12 text-gray-900">役職紹介</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 lg:gap-8">
            <div class="bg-white p-8 rounded-xl shadow-lg border border-gray-100 hover:-translate-y-2 hover:shadow-2xl transition-all duration-300 text-center">
                <h3 class="text-xl md:text-2xl font-bold mb-4 text-[#4a69bd]">行事委員長</h3>
                <p class="text-gray-700 leading-relaxed">委員会全体のリーダーとして、活動をまとめます。</p>
            </div>
            <div class="bg-white p-8 rounded-xl shadow-lg border border-gray-100 hover:-translate-y-2 hover:shadow-2xl transition-all duration-300 text-center">
                <h3 class="text-xl md:text-2xl font-bold mb-4 text-[#4a69bd]">副行事委員長</h3>
                <p class="text-gray-700 leading-relaxed">委員長を補佐し、円滑な運営をサポートします。</p>
            </div>
            <div class="bg-white p-8 rounded-xl shadow-lg border border-gray-100 hover:-translate-y-2 hover:shadow-2xl transition-all duration-300 text-center">
                <h3 class="text-xl md:text-2xl font-bold mb-4 text-[#4a69bd]">会計</h3>
                <p class="text-gray-700 leading-relaxed">委員会の予算を管理し、活動資金を確保します。</p>
            </div>
            <div class="bg-white p-8 rounded-xl shadow-lg border border-gray-100 hover:-translate-y-2 hover:shadow-2xl transition-all duration-300 text-center">
                <h3 class="text-xl md:text-2xl font-bold mb-4 text-[#4a69bd]">書記</h3>
                <p class="text-gray-700 leading-relaxed">会議の議事録を作成し、活動の記録を残します。</p>
            </div>
        </div>
    </div>
</section>

<!-- Joinセクション -->
<section id="join" class="py-16 md:py-20 bg-gradient-to-r from-[#4a69bd] to-[#6a89cc] text-white">
    <div class="container mx-auto px-4 sm:px-6 lg:px-8">
        <div class="max-w-4xl mx-auto text-center">
            <h2 class="text-3xl md:text-4xl font-bold mb-6">君も仲間にならないか？</h2>
            <div class="space-y-4 text-lg md:text-xl leading-relaxed opacity-95">
                <p>行事委員会は、各クラスから2名まで参加できます。</p>
                <p>学校を盛り上げたい、新しいことに挑戦したい、そんな熱い想いを持った君を待っています！</p>
                <p>興味のある方は、来年に行事委員会に立候補してみてください。</p>
            </div>
        </div>
    </div>
</section>