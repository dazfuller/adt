# Azure Digital Twin command line utility

This is a small command line application utility I wrote to assist with model management of Azure Digital Twin instances, whilst diving deeper into the [Go](https://go.dev/) programming language.

## So, why not just use the existing CLI?

The Azure Digital Twins Models API for [adding models](https://docs.microsoft.com/rest/api/digital-twins/dataplane/models/digitaltwinmodels_add) doesn't state it in the documentation, but it doesn't really have the concept of a batch. That is, if you have more models than the 250 model maximum limit per API request, there isn't a way of performing multiple uploads and flagging them as being part of the same batch. If you upload up to 40 at a time then the service does interpret this as a batch, but still assumes that each API request is complete in itself.

Because of this the other thing which isn't really documented can bite you. Models _**must**_ be uploaded in dependency order. So, if you have a model in api call 1 which depends on a model in api call 3, then the first batch will fail.

The [Azure CLI](https://docs.microsoft.com/azure/digital-twins/concepts-cli) does a better job of handling this than calling the API directly because it groups the models based on their dependency count and then [uploads](https://github.com/Azure/azure-iot-cli-extension/blob/67d6c2aa7414f89be25abf26671b88192ab13bee/azext_iot/digitaltwins/providers/model.py#L73) based on dependency count order (so models with 0 dependencies go first, then 1 dependency etc...). This is better but it's not perfect as we could have a model with a single dependency, but that dependent model has 5 of it's own.

The same is also true of deleting all models from an instance. You can _**not**_ delete a model which has other models depending on it.

## What does this do differently?

This code implements a [topological sort](https://wikipedia.org/wiki/Topological_sorting) to ensure that models are only part of an API request when all of their dependencies (and their dependent dependencies and so on) have either been uploaded already or are part of the same request. And when clearing all models it does the same, ensuring that a model is only deleted when it's dependent graph is fully deleted.
