//! Test suite for the Web and headless browsers.

#![cfg(target_arch = "wasm32")]

extern crate wasm_bindgen_test;
use wasm::greet;
use wasm_bindgen_test::console_log;
use wasm_bindgen_test::*;

wasm_bindgen_test_configure!(run_in_browser);

#[wasm_bindgen_test]
fn pass() {
    assert_eq!(1 + 1, 2);
}

#[wasm_bindgen_test]
async fn simple_example() {
    let res = reqwest::get("https://hyper.rs")
        .await
        .expect("http get example");
    console_log!("Status: {}", res.status());

    let body = res.text().await.expect("response to utf-8 text");
    console_log!("Body:\n\n{}", body);
}


const RSSHUB_TEST_CONTENT: &'static str = r#"
<rss version="2.0">
<channel>
<title>3DM - 新闻中心</title>
<link>http://www.3dmgame.com/news/</link>
<atom:link href="http://rsshub.app/3dm/news" rel="self" type="application/rss+xml"/>
<description>
3DM - 新闻中心 - Made with love by RSSHub(https://github.com/DIYgod/RSSHub)
</description>
<generator>RSSHub</generator>
<webMaster>i@diygod.me (DIYgod)</webMaster>
<language>zh-cn</language>
<lastBuildDate>Fri, 15 Apr 2022 02:02:48 GMT</lastBuildDate>
<ttl>120</ttl>
<item>
<title>CD Projekt：《巫师3》次世代升级绝未处于“开发地狱”</title>
<description>
<p style="text-indent:2em;"> CD Projekt 昨天宣布，原计划于 2020 年首次宣布的《巫师 3：狂猎》的次世代主机升级计划再次推迟。该消息并未附带正式发布日期，只是表示升级已被推迟“直至另行通知”。 </p> <p align="center"> <img src="https://img.3dmgame.com/uploads/images/news/20220415/1649987387_937083.jpg" alt="CD Projekt：《巫师3》次世代升级绝未处于“开发地狱”" referrerpolicy="no-referrer"> </p> <p style="text-indent:2em;"> 对于许多熟悉游戏开发过程的玩家来说，“直到另行通知”听起来很像“无限期延迟”，这通常是陷入严重麻烦的游戏的信号。像是《吸血鬼：避世血族-血统2》在 2021 年初宣布将被无限期推迟，最终导致了预售终止，开发商最终取消了该项目。除了模糊地表示开发仍在继续以外，从那以后再也没有任何其他消息。 </p> <p style="text-indent:2em;"> 然而，在今天的投资者电话会议中，CD Projekt 的业务发展高级副总裁 Michał Nowakowski 指出，备受期待的《巫师 3》升级并非如此。 </p> <p style="text-indent:2em;"> “我一直在关注互联网上所有的新闻，我发现了一个真正引起我注意的头条新闻，那就是‘《巫师3》次世代无限期延迟’，这听起来像是游戏在某种开发地狱中，”他说道：“我想说这不是事实。我们将推出很多预告，比如明年 6 月或相近的时间。完全不是（像有些新闻陈述的）那样。” </p> <p align="center"> <img src="https://img.3dmgame.com/uploads/images/news/20220415/1649987412_238870.jpg" alt="CD Projekt：《巫师3》次世代升级绝未处于“开发地狱”" referrerpolicy="no-referrer"> </p> <p style="text-indent:2em;"> “我们说的全部都是，我们已经在内部进行了游戏的开发。游戏将在内部完成。我们正在评估我们的时间，这需要一些调查。这就是我们所说的一切。没有人说游戏被推迟到很长的时间之后。这就是我现在关于《巫师》次世代所能说的，但我真的很想强调这一事实。” </p> <p style="text-indent:2em;"> 在电话会议中，CD Projekt 联合首席执行官 Adam Kicinski 也表示，在升级方面没有大量工作要做：“（剩下的工作）不多。”当被问及完成该项目需要多少开发人员时，他表示估计总数约为 15 人，外加一些外部支持人员：“接手这个项目不会影响我们下一款游戏的开发。” </p> <p style="text-indent:2em;"> 《巫师3》次世代升级正在为 PS5 和 Xbox Series X|S 游戏机开发，并将作为 PC 版游戏的免费更新。电话会议期间没有透露可能的新发布日期，但可能并不像猜测的那样遥远。不过电话会议中也提到了一个坏消息：《赛博朋克2077》的扩展要到 2023 年某个时候才会推出。 </p>
</description>
<pubDate>Fri, 15 Apr 2022 01:50:00 GMT</pubDate>
<guid isPermaLink="false">https://www.3dmgame.com/news/202204/3840341.html</guid>
<link>https://www.3dmgame.com/news/202204/3840341.html</link>
<category>巫师3：狂猎</category>
</item>
<item>
<title>科乐美社交对战推理游戏《CRIMESIGHT》现已发售 支持中文</title>
<description>
<p style="text-indent:2em;"> 由科乐美开发的社交对战推理游戏《CRIMESIGHT》现已在Steam发售，售价19.99美元，目前锁国区，但支持简体中文。 </p> <p style="text-indent:2em;"> 游戏中，玩家不仅要推理出凶手，还要推理出受害者！？令人耳目一新的推理模拟游戏登场！6名男女被困在暴风雪中的洋房里。已经预测到他们中的一人会杀害另外一人。 目前，不仅凶手的身份不明，连受害者会是谁也不得而知。“CRIMESIGHT”是一款多人对战推理游戏。玩家分为两个阵营，一方要和刑侦AI“Sherlock”一 </p> <h3> 游戏预告 </h3> <p align="center"> <iframe height="480" width="640" src="https://player.youku.com/embed/XNTg2MTU3NzQ1Ng?client_id=5a73c0df8eb0d91d" frameborder="0" allowfullscreen=""> </iframe> </p> <p style="text-indent:2em;"> <strong>游戏Steam页面：</strong><a href="https://store.steampowered.com/app/1491710/CRIMESIGHT/" target="_blank"><strong>点击此处</strong></a><strong>（锁国区）</strong> </p> <h3> 游戏简介 </h3> <p style="text-indent:2em;"> 不仅要推理出凶手，还要推理出受害者！？令人耳目一新的推理模拟游戏登场！ </p> <p style="text-indent:2em;"> 2075年的伦敦。 </p> <p style="text-indent:2em;"> 根据网络上的所有信息预测未来恶性犯罪的综合系统：“Foresight AI”已投入实用。此系统的实用减少了世界的恶性犯罪90%。 </p> <p style="text-indent:2em;"> 然而，“Foresight AI”预测到，即便用“Foresight AI”进行预测也会出现无法预防的事件，而该事件会导致世界走向灭亡。 </p> <p style="text-indent:2em;"> 系统的开发者担心会一语成谶，于是让系统派生出特殊AI，以追踪、解决这种尚未发生的恶性犯罪。 </p> <p style="text-indent:2em;"> 这个AI名叫“Sherlock”。这是登场于旧时代小说中的绝世名侦探的名字。 </p> <p style="text-indent:2em;"> 在追踪众多案件的过程中，“Sherlock”推导出一个事实。 </p> <p style="text-indent:2em;"> 一个能力足以与自己相匹敌的犯罪策划AI——“Moriarty”如同一只端坐网中的蜘蛛，与案件有着千丝万缕的联系。 </p> <p align="center"> <img src="https://cdn.cloudflare.steamstatic.com/steam/apps/1491710/extras/stage.gif?t=1649961679" referrerpolicy="no-referrer"> </p> <h3> 游戏特色 </h3> <p style="text-indent:2em;"> 暴风雪中的洋房……受害者不明的杀人案即将发生？！ </p> <p style="text-indent:2em;"> 游戏中的角色分为实施凶杀的“杀手”和被杀手盯上的“目标”。 </p> <p style="text-indent:2em;"> 需要满足几个条件才能成功实施凶杀，Moriarty方的玩家必须满足那些条件。而Sherlock方的玩家则必须阻止其杀人。 </p> <p style="text-indent:2em;"> 在双方较量的过程中，“杀手”和“目标”的身份将逐渐明朗。 </p> <p align="center"> <img src="https://cdn.cloudflare.steamstatic.com/steam/apps/1491710/extras/detective.gif?t=1649961679" referrerpolicy="no-referrer"> </p> <p style="text-indent:2em;"> 你能和Moriarty合作，瞒过Sherlock完成杀害目标的计划吗？ </p> <p style="text-indent:2em;"> 或者，能和Sherlock合作，阻止Moriarty的计划吗？ </p> <h3> 游戏截图 </h3> <p align="center"> <img src="https://img.3dmgame.com/uploads/images/news/20220415/1649986709_660059.png" alt="科乐美社交对战推理游戏《CRIMESIGHT》现已发售 支持中文" referrerpolicy="no-referrer"> </p> <p align="center"> <img src="https://img.3dmgame.com/uploads/images/news/20220415/1649986718_822697.png" alt="科乐美社交对战推理游戏《CRIMESIGHT》现已发售 支持中文" referrerpolicy="no-referrer"> </p> <p align="center"> <img src="https://img.3dmgame.com/uploads/images/news/20220415/1649986724_460267.png" alt="科乐美社交对战推理游戏《CRIMESIGHT》现已发售 支持中文" referrerpolicy="no-referrer"> </p> <p align="center"> <img src="https://img.3dmgame.com/uploads/images/news/20220415/1649986775_114882.png" alt="科乐美社交对战推理游戏《CRIMESIGHT》现已发售 支持中文" referrerpolicy="no-referrer"> </p> <p align="center"> <img src="https://img.3dmgame.com/uploads/images/news/20220415/1649986814_469752.png" alt="科乐美社交对战推理游戏《CRIMESIGHT》现已发售 支持中文" referrerpolicy="no-referrer"> </p> <h3> 系统需求 </h3> <p style="text-indent:2em;"> 最低配置 </p> <p style="text-indent:2em;"> 需要64位处理器和操作系统 </p> <p style="text-indent:2em;"> 操作系统: Windows10(64bit) </p> <p style="text-indent:2em;"> 处理器: Intel Core i5 7300U 2.6GHz </p> <p style="text-indent:2em;"> 内存: 8 GB RAM </p> <p style="text-indent:2em;"> 显卡: Intel HD Graphics 620 </p> <p style="text-indent:2em;"> DirectX版本: 11 </p> <p style="text-indent:2em;"> 存储空间:需要2 GB可用空间 </p>
</description>
<pubDate>Fri, 15 Apr 2022 01:45:00 GMT</pubDate>
<guid isPermaLink="false">https://www.3dmgame.com/news/202204/3840340.html</guid>
<link>https://www.3dmgame.com/news/202204/3840340.html</link>
<category>CRIMESIGHT</category>
</item>
</channel>
</rss>
"#;