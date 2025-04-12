<script lang="ts">
    let { children } = $props();

</script>

<div style="display:contents;" class="button-wrapper" onclick={(e)=>{e.stopPropagation(); e.preventDefault()}}>
{@render children()}
</div>