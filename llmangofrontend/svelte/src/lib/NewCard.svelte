<script>
    let {text, onclick}=$props()
    import Card from "./Card.svelte";

</script>
<button class="wrapper" onclick={onclick}>
<Card>
    + {text}
</Card>
</button>

<style>
    .wrapper{
        background: none;
        border: none;
        padding: 0;
        margin: 0;
        cursor: pointer;
        width: 100%;
        text-align: left;
    }
</style>