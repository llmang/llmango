<script lang="ts">
    import { untrack } from "svelte";

  type Props = {
    id: string;
    label: string;
    value: string | number | undefined;
    children?: () => any;
  };
  let { id, label, value=$bindable(), children } : Props = $props();
</script>

<div class="filter-select-wrapper">
  {#if value !== undefined && value !== null && value !== ''}
    <div class="floating-label">{label}</div>
  {/if}

  <select id={id} bind:value>
    {@render children?.()}
  </select>
</div>

<style>
  .filter-select-wrapper {
    position: relative;
    display: flex;
    max-width: 100%;
  }
  .floating-label {
    position: absolute;
    top: -1.25rem;
    font-weight: 800;
    color: lightgrey;
    font-size: 0.8rem;
    pointer-events: none;
  }

  select {
  /* Reset defaults */
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-color: #fff;
  color: #333;
  border: 1.5px solid #ccc;
  padding: .25em 1.5em .25em .5em;
  border-radius: 0.5em;
  font-size: 1rem;
  font-family: inherit;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  cursor: pointer;
  box-shadow: 0 1px 6px rgba(60,60,80,0.04);
  width: 10rem;
  max-width: 100%;

  /* Space for arrow */
  background-position: right .1em center;
  background-repeat: no-repeat;
  /* SVG chevron as data URI */
  background-image:
    url('data:image/svg+xml;charset=UTF-8,<svg fill="gray" height="20" viewBox="0 0 24 24" width="20" xmlns="http://www.w3.org/2000/svg"><path d="M7.41 8.59 12 13.17l4.59-4.58L18 10l-6 6-6-6z"/></svg>');
}

select:focus {
  border-color: #54a9ff;
  box-shadow: 0 0 0 2px #aacfff99;
}

/* Remove background image on IE */
select::-ms-expand {
  display: none;
}
</style> 