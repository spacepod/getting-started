/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package com.google.cloud.genomics.gettingstarted;

import com.beust.jcommander.JCommander;
import com.beust.jcommander.Parameter;
import com.beust.jcommander.ParameterException;
import com.beust.jcommander.internal.Maps;
import com.google.api.client.extensions.java6.auth.oauth2.VerificationCodeReceiver;
import com.google.api.client.extensions.jetty.auth.oauth2.LocalServerReceiver;
import com.google.api.client.googleapis.extensions.java6.auth.oauth2.GooglePromptReceiver;
import com.google.api.services.genomics.Genomics;
import com.google.api.services.genomics.GenomicsScopes;
import com.google.api.services.genomics.model.Read;
import com.google.api.services.genomics.model.Readset;
import com.google.api.services.genomics.model.SearchReadsRequest;
import com.google.api.services.genomics.model.SearchReadsetsRequest;
import com.google.cloud.genomics.utils.GenomicsFactory;
import com.google.common.base.Suppliers;
import com.google.common.collect.Lists;

import java.io.File;
import java.io.IOException;
import java.math.BigInteger;
import java.util.List;
import java.util.Map;

/**
 * Getting started with Google Genomics in Java
 */
public class MainExample {

  private static class Arguments {
    @Parameter(names = "--client_secrets_filename",
        description = "Path to client_secrets.json")
    public String clientSecretsFilename = "client_secrets.json";

    @Parameter(names = "--noauth_local_webserver",
        description = "If your browser is on a different machine then run this" +
            "application with this command-line parameter")
    public boolean noLocalServer = false;
  }

  public static void main(String[] args) throws IOException {
    Arguments arguments = new Arguments();
    JCommander parser = new JCommander(arguments);

    try {
      // Parse the command line
      parser.parse(args);

      // Authorization
      VerificationCodeReceiver receiver = arguments.noLocalServer ?
          new GooglePromptReceiver() : new LocalServerReceiver();
      GenomicsFactory genomicsFactory = GenomicsFactory.builder("getting_started_java")
          .setScopes(Lists.newArrayList(GenomicsScopes.GENOMICS))
          .setVerificationCodeReceiver(Suppliers.ofInstance(receiver))
          .build();

      File clientSecrets = new File(arguments.clientSecretsFilename);
      if (!clientSecrets.exists()) {
        System.err.println(
            "Client secrets file " + arguments.clientSecretsFilename + " does not exist."
            + " Visit https://developers.google.com/genomics to learn how"
            + " to install a client_secrets.json file.  If you have installed a client_secrets.json"
            + " in a specific location, use --client_secrets_filename <path>/client_secrets.json.");
        return;
      }
      Genomics genomics = genomicsFactory.fromClientSecretsFile(clientSecrets);


      //
      // This example gets the read bases for NA12878 at specific a position
      //
      String datasetId = "376902546192"; // This is the 1000 Genomes dataset ID
      String referenceName = "22";
      final Integer referencePosition = 51005354;

      // 1. First find the readset ID for NA12878
      SearchReadsetsRequest readsetsReq = new SearchReadsetsRequest()
          .setDatasetIds(Lists.newArrayList(datasetId))
          .setName("NA12878");

      List<Readset> readsets = genomics.readsets().search(readsetsReq)
          .setFields("readsets(id)").execute().getReadsets();
      if (readsets == null || readsets.size() != 1) {
        System.err.println("Searching for NA12878 didn't return the right number of results");
        return;
      }

      String na12878 = readsets.get(0).getId();


      // 2. Once we have the readset ID,
      // lookup the reads at the position we are interested in
      SearchReadsRequest readsReq = new SearchReadsRequest()
          .setReadsetIds(Lists.newArrayList(na12878))
          .setSequenceName(referenceName)
          .setSequenceStart(BigInteger.valueOf(referencePosition))
          .setSequenceEnd(BigInteger.valueOf(referencePosition))
          .setMaxResults(BigInteger.valueOf(1024));

      List<Read> reads = genomics.reads().search(readsReq)
          .setFields("reads(position,originalBases,cigar)").execute().getReads();

      Map<Character, Integer> baseCounts = Maps.newHashMap();
      for (Read read : reads) {
        int index = referencePosition - read.getPosition();
        // Note: This is simplistic - the cigar should be considered for real code
        Character base = read.getOriginalBases().charAt(index);

        if (!baseCounts.containsKey(base)) {
          baseCounts.put(base, 0);
        }
        baseCounts.put(base, baseCounts.get(base) + 1);
      }

      System.out.println("NA12878 bases on " + referenceName + " at " + referencePosition);
      for (Map.Entry<Character, Integer> entry : baseCounts.entrySet()) {
        System.out.println(entry.getKey() + ": " + entry.getValue());
      }


    } catch (ParameterException e) {
      System.err.append(e.getMessage()).append("\n");
      parser.usage();
    } catch (IllegalStateException e) {
      System.err.println(e.getMessage());
    } catch (Throwable t) {
      t.printStackTrace();
    }
  }
}
